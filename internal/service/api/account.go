package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hisshihi/simple-bank/db/sqlc"
)

type createAccountRequest struct {
	Owner    string `json:"owner" binding:"required"`
	Currency string `json:"currency" binding:"required,oneof=USD RUB EUR"`
}

type createAccountResponse struct {
	Owner    string `json:"owner"`
	Currency string `json:"currency"`
	Balance  int64  `json:"balance"`
}

func (server *Server) createAccount(ctx *gin.Context) {
	var req createAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := sqlc.CreateAccountParams{
		Owner:    req.Owner,
		Balance:  0,
		Currency: req.Currency,
	}

	account, err := server.store.CreateAccount(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := createAccountResponse{
		Owner:    account.Owner,
		Currency: account.Currency,
		Balance:  account.Balance,
	}

	ctx.JSON(http.StatusOK, rsp)
}
