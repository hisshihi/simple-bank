package api

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hisshihi/simple-bank/db/sqlc"
)

type transferRequest struct {
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64  `json:"to_account_id" binding:"required,min=1"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Currency      string `json:"currency" binding:"required,oneof=USD RUB EUR"`
}

type transferResponse struct {
	ID       int64  `json:"id"`
	Owner    string `json:"owner"`
	Currency string `json:"currency"`
	Balance  int64  `json:"balance"`
}

func (server *Server) createTransfer(ctx *gin.Context) {
	var req transferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if !server.validAccount(ctx, req.FromAccountID, req.Currency) {
		return
	}

	if !server.validAccount(ctx, req.ToAccountID, req.Currency) {
		return
	}

	if !server.validAmount(ctx, req.FromAccountID, req.Amount) {
		return
	}

	arg := sqlc.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}

	result, err := server.store.TransferTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (server *Server) validAmount(ctx *gin.Context, accountID int64, amount int64) bool {
	account, err := server.store.GetAccount(ctx, accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return false
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return false
	}

	if account.Balance < amount {
		err = fmt.Errorf("недостаточно средств для перевода: %d", account.Balance)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "недостаточно средств для перевода"})
		return false
	}
	return true
}

func (server *Server) validAccount(ctx *gin.Context, accountID int64, currency string) bool {
	account, err := server.store.GetAccount(ctx, accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return false
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return false
	}

	if account.Currency != currency {
		err = fmt.Errorf("аккаунт: [%d] не соответствие валют %v и %v", account.ID, account.Currency, currency)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "не соответствие валют аккаунта"})
		return false
	}

	return true
}
