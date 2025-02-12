package sqlc

import (
	"context"
	"database/sql"
	"fmt"
)

// Store предоставляет все функции для доступа к базе данных а также транзакции
type Store struct {
	*Queries
	db *sql.DB
}

// NewStore создает новый Store
func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

// execTx выполняет функцию в транзакции и откатывается при ошибке
func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
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
func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	/*
		Создаём транзакцию, в которой будут выполнены все операции
		Если какая-то операция не выполнится, то транзакция будет отменена
	*/
	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		// Создаём запись в истории транзакций
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams(arg))
		if err != nil {
			return err
		}

		// Создаём запись в истории транзакций для счета отправителя
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}

		// Создаём запись в истории транзакций для счета получателя
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		// TODO: обновить балансы

		return nil
	})
	return result, err
}
