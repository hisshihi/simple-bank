package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hisshihi/simple-bank/db/sqlc"
)

type createEntryRequest struct {
	AccountID int64 `json:"account_id" binding:"required,min=1"`
	Amount    int64 `json:"amount" binding:"required"`
}

type entryResponse struct {
	AccountID int64 `json:"account_id"`
	Amount    int64 `json:"amount" `
}

func (server *Server) createEntry(ctx *gin.Context) {
	var req createEntryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := sqlc.CreateEntryParams{
		AccountID: req.AccountID,
		Amount:    req.Amount,
	}

	entry, err := server.store.CreateEntry(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := entryResponse{
		AccountID: entry.AccountID,
		Amount:    entry.Amount,
	}

	ctx.JSON(http.StatusOK, rsp)
}

type getEntryRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getEntry(ctx *gin.Context) {
	var req getEntryRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	entry, err := server.store.GetEntry(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := entryResponse{
		AccountID: entry.AccountID,
		Amount:    entry.Amount,
	}

	ctx.JSON(http.StatusOK, rsp)
}

type listEntriesRequest struct {
	PageID    int32 `form:"page_id" binding:"required,min=1"`
	PageSize  int32 `form:"page_size" binding:"required,min=5,max=10"`
	AccountID int64 `form:"account_id" binding:"required,min=1"`
}

func (server *Server) listEntries(ctx *gin.Context) {
	var req listEntriesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := sqlc.ListEntriesParams{
		Limit:     int64(req.PageSize),
		Offset:    int64((req.PageID - 1) * req.PageSize),
		AccountID: req.AccountID,
	}

	entries, err := server.store.ListEntries(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, entries)
}
