package api

import (
	"database/sql"
	"fmt"

	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/karlib/simple_bank/db/sqlc"
)

type createAccountRequest struct {
	Owner    string `json:"owner" binding:"required"`
	Currency string `json:"currency" binding:"required,oneof=USD EUR"`
}

func (server *Server) createAccount(ctx *gin.Context) {
	var req createAccountRequest
	// pokud error není nil klient poskytl nesprávné údaje
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreateAccountParams{
		Owner:    req.Owner,
		Currency: req.Currency,
		Balance:  0,
	}

	account, err := server.store.CreateAccount(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
	}

	ctx.JSON(http.StatusOK, account)
}

type getAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getAccountByID(ctx *gin.Context) {
	var req getAccountRequest
	// pokud error není nil klient poskytl nesprávné údaje
	if err := ctx.ShouldBindUri(&req); err != nil {
		// http.StatusBadRequest is code 400
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	account, err := server.store.GetAccount(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			// if we get an error type ErrNoRows server will respond with code 404 (not found)
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, account)
}

// form tag zařídí, že se hodnoty do reqestu dostanout z QueryParam, page size má nadefinované tagy min a max pro rozmezí
// povolené požadované velikosti stránky s výsledky
type listAccountRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=1,max=10"`
}

func (server *Server) listAccount(ctx *gin.Context) {
	var req listAccountRequest
	// ShouldBindQuery řekne GIN frameworku, aby vzal query data z requestu
	if err := ctx.ShouldBindQuery(&req); err != nil {
		// http.StatusBadRequest is code 400
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	fmt.Println((req.PageID - 1) * req.PageSize)
	arg := db.ListAccountsParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	accounts, err := server.store.ListAccounts(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, accounts)
}
