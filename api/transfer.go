package api

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/ulunnuha-h/simple_bank/db/sqlc"
)

type createTransferRequest struct {
	FromAccountId int64 `json:"from_account_id" binding:"required,min=1"`
	ToAccountId   int64 `json:"to_account_id" binding:"required,min=1"`
	Amount        int64 `json:"amount" binding:"required,gt=0"`
	Currency string `json:"currency" binding:"required,currency"`
}

func (server *Server) createTransfer(ctx *gin.Context) {
	var req createTransferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload, err := GetAuthPayload(ctx)
	if err != nil {
		return
	}

	if !server.validateAccount(ctx, req.FromAccountId, req.Currency, req.Amount, true, authPayload.Username) ||
		!server.validateAccount(ctx, req.ToAccountId, req.Currency, 0, false, "") {
		return
	}

	args := db.TransferTxParams{
		FromAccountId: req.FromAccountId,
		ToAccountId:   req.ToAccountId,
		Amount:        req.Amount,
	}

	result, err := server.store.TransferTx(ctx, args)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (server *Server) validateAccount(ctx *gin.Context, accountID int64, currency string, amount int64, checkBalance bool, username string) bool {
	account, err := server.store.GetAccount(ctx, accountID)
	if err != nil {
		status := http.StatusInternalServerError
		if err == sql.ErrNoRows {
			status = http.StatusNotFound
		}
		ctx.JSON(status, errorResponse(err))
		return false
	}

	if(account.Owner != username && checkBalance) {
		ctx.JSON(http.StatusForbidden, errorResponse(fmt.Errorf("user [%s] does not have access to account with ID %d", username, accountID)))
		return false
	}

	if account.Currency != currency {
		ctx.JSON(http.StatusBadRequest, errorResponse(fmt.Errorf("account [%d] currency mismatch: %s and %s", accountID, account.Currency, currency)))
		return false
	}

	if checkBalance && account.Balance < amount {
		ctx.JSON(http.StatusBadRequest, errorResponse(fmt.Errorf("account [%d] has insufficient balance: required %d, available %d", accountID, amount, account.Balance)))
		return false
	}

	return true
}
