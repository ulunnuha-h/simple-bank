package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	db "github.com/ulunnuha-h/simple_bank/db/sqlc"
	"github.com/ulunnuha-h/simple_bank/token"
)

type createAccountRequest struct {
	Currency string `json:"currency" binding:"required,oneof=USD EUR IDR"`
}

func (server *Server) createAccount(ctx *gin.Context){
	var	req createAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil{
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload, err := GetAuthPayload(ctx)
	if err != nil {
		return
	}

	args := db.CreateAccountParams{
		Owner: authPayload.Username,
		Currency: req.Currency,
		Balance: 0,
	}

	account, err := server.store.CreateAccount(ctx, args)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "foreign_key_violation", "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return	
			}
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, account)
}

type getAccountRequest struct{
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getAccount(ctx *gin.Context){
	var req getAccountRequest
	if err := ctx.ShouldBindUri(&req); err != nil{
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload, err := GetAuthPayload(ctx)
	if err != nil {
		return
	}

	account, err := server.store.GetAccount(ctx, req.ID)
	if err != nil {
		if(err == sql.ErrNoRows){
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if(account.Owner != authPayload.Username) {
		ctx.JSON(http.StatusForbidden, errorResponse(token.ErrDoesNotBelong))
		return
	}

	ctx.JSON(http.StatusOK, account)
}

type listAccountRequest struct{
	PAGE_ID int32 `form:"page_id" binding:"required,min=1"`
	PAGE_SIZE int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) listAccount(ctx *gin.Context){
	var req listAccountRequest
	if err := ctx.ShouldBindQuery(&req); err != nil{
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload, err := GetAuthPayload(ctx)
	if err != nil {
		return
	}

	args := db.ListAccountsParams{
		Owner: authPayload.Username,
		Limit: req.PAGE_SIZE,
		Offset: (req.PAGE_ID - 1) * req.PAGE_SIZE,
	}

	accounts, err := server.store.ListAccounts(ctx, args)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, accounts)
}

type deleteAccountRequest struct{
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) deleteAccount(ctx *gin.Context){
	var req deleteAccountRequest
	if err := ctx.ShouldBindUri(&req); err != nil{
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload, err := GetAuthPayload(ctx)
	if err != nil {
		return
	}

	account, err := server.store.GetAccount(ctx, req.ID)
	if err != nil {
		if(err == sql.ErrNoRows){
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if(account.Owner != authPayload.Username){
		ctx.JSON(http.StatusForbidden, errorResponse(token.ErrActionForbidden))
		return
	}

	err = server.store.DeleteAccount(ctx, req.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message":"Deleted Succesfully!"})
}

type updateAccountUriRequest struct{
	ID int64 `uri:"id" binding:"required,min=1"`
}

type updateAccountJsonRequest struct{
	Balance int64 `json:"balance" binding:"required"`
}

func (server *Server) updateAccount(ctx *gin.Context){
	var req_uri updateAccountUriRequest
	var req_json updateAccountJsonRequest

	if err := ctx.ShouldBindUri(&req_uri); err != nil{
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if err := ctx.ShouldBindJSON(&req_json); err != nil{
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	args := db.UpdateAccountParams{
		ID: req_uri.ID,
		Balance: req_json.Balance,
	}

	account, err := server.store.UpdateAccount(ctx, args)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, account)
}