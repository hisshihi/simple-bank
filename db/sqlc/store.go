package sqlc

import (
	"context"
	"database/sql"
	"fmt"
)

// Store - это хранилище будет предоставлять все функции для выполнения запросов к базе данных по отдельности, а также их комбинации в рамках транзакции.
type Store struct {
	// Queries - это вспомогательный объект, который содержит все запросы к базе данных.
	*Queries
	// db - необходим для создания новой транзакции.
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Создаём новый объект Queries, который будет использоваться для выполнения запросов к базе данных внутри транзакции.
	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"` // ID счета, с которого будет перевод
	ToAccountID   int64 `json:"to_account_id"`   // ID счета, на который будет перевод
	Amount        int64 `json:"amount"`          // Сумма перевода
}

type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`     // Созданная запись перевода
	FromAccount Account  `json:"from_account"` // Поле вычитания средств
	ToAccount   Account  `json:"to_account"`   // Поле прибавления средств
	FromEntry   Entry    `json:"from_entry"`   // Созданная запись, что деньги перемещаются из FromAccount
	ToEntry     Entry    `json:"to_entry"`     // Созданная запись, что деньги перемещаются в ToAccount
}

// TransferTx выполняет транзакцию перевода денег из одного счета на другой.
// Функция переводит деньги из одного счёта на другой, создаёт записи о переводе и обновляет балансы счетов.
func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams(arg))
		if err != nil {
			return err
		}

		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}

		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		// С помощью условия определяем аккаунт с меньшим идентификатором, чтобы избежать deadlock
		if arg.FromAccountID < arg.ToAccountID {
			result.FromAccount, result.ToAccount, err = addMoney(ctx, q, arg.FromAccountID, -arg.Amount, arg.ToAccountID, arg.Amount)
			if err != nil {
				return err
			}
		} else {
			result.ToAccount, result.FromAccount, err = addMoney(ctx, q, arg.ToAccountID, arg.Amount, arg.FromAccountID, -arg.Amount)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return result, err
}

func addMoney(
	ctx context.Context,
	q *Queries,
	accountID1 int64,
	amount1 int64,
	accountID2 int64,
	amount2 int64,
) (account1, account2 Account, err error) {
	account1, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID: accountID1,
		Amount: amount1,
	})
	if err != nil {
		return
	}

	account2, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID: accountID2,
		Amount: amount2,
	})
	if err != nil {
		return
	}

	return
}