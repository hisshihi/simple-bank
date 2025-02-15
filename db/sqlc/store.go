package sqlc

import (
	"context"
	"database/sql"
	"fmt"
)

type Store interface {
	Querier
	TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error)
}

/*
Объединяет SQL-запросы (сгенерированные sqlc) и подключение к БД.

	Это центральная точка для выполнения операций.
*/
type SQLStore struct {
	*Queries
	db *sql.DB
}

// NewStore создает новый Store
func NewStore(db *sql.DB) Store {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}

// execTx выполняет функцию в транзакции и откатывается при ошибке
func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil) // BeginTx начинает транзакцию, nil - уровень изоляции(по умолчанию)
	if err != nil {
		return err
	}

	q := New(tx) // создаем новый Queries с tx, tx - база данных
	err = fn(q)  // выполняем функцию fn, fn - функция, которая будет выполнена в транзакции
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil { // откат транзакции при ошибке
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit() // коммит транзакции
}

// TransferTxParams содержит все необходимые параметры для перевода денег
type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

// TransferTxResult содержит результат перевода денег
type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

// TransferTx переводит деньги из одного счета на другой
// Создаст новую запись в истории транзакций, добавит сумму на счет и вычтет сумму с другого счета
func (store *SQLStore) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	/*
		Создаём транзакцию, в которой будут выполнены все операции
		Если какая-то операция не выполнится, то транзакция будет отменена
	*/
	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		// Создание записи о переводе
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams(arg))
		if err != nil {
			return err
		}

		// Создание двух записей в истории операций (дебет и кредит)
		// Создаём запись в истории транзакций для счета отправителя
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}

		// Создание двух записей в истории операций (дебет и кредит)
		// Создаём запись в истории транзакций для счета получателя
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		// Обновление балансов счетов с deadlock prevention:
		// Если id счета отправителя меньше id счета получателя, то переводим деньги от отправителя к получателю
		// Иначе переводим деньги от получателя к отправитлю

		/*
			Ключевой момент: порядок обновления счетов всегда определяется их ID.
			 Это предотвращает deadlock'и при параллельных операциях.
		*/
		if arg.FromAccountID < arg.ToAccountID {
			result.FromAccount, result.ToAccount, err = addMoney(ctx, q, arg.FromAccountID, -arg.Amount, arg.ToAccountID, arg.Amount)
		} else {
			result.ToAccount, result.FromAccount, err = addMoney(ctx, q, arg.ToAccountID, arg.Amount, arg.FromAccountID, -arg.Amount)
		}

		return nil
	})
	return result, err
}

/*
Выполняет атомарное обновление балансов
Всегда обновляет счета в строгом порядке (по возрастанию ID)
*/
func addMoney(
	ctx context.Context,
	q *Queries,
	accountID1 int64,
	amount1 int64,
	accountID2 int64,
	amount2 int64,
) (account1 Account, account2 Account, err error) {
	account1, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID1,
		Amount: amount1,
	})
	if err != nil {
		return
	}

	account2, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID2,
		Amount: amount2,
	})
	if err != nil {
		return
	}

	return
}