package api

import (
	"database/sql"
	"errors"

	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/karlib/simple_bank/db/sqlc"
	"github.com/karlib/simple_bank/token"
	"github.com/lib/pq"
)

type createAccountRequest struct {
	Currency string `json:"currency" binding:"required,currency"`
}

func (server *Server) createAccount(ctx *gin.Context) {
	var req createAccountRequest
	// pokud error není nil klient poskytl nesprávné údaje
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// podle kliče se kterým v middlewaru uložím do kontextu hondotu, ji zde
	// vytáhnu ven a dál s ní pracuji
	// tahle operace vraci general interface takže je nutné to castnout na správný typ pomocí .(*token.Payload)
	// to dostanu rovnou ve formatu token.Payload
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	arg := db.CreateAccountParams{
		Owner:    authPayload.Username,
		Currency: req.Currency,
		Balance:  0,
	}

	account, err := server.store.CreateAccount(ctx, arg)
	if err != nil {
		// zde je třeba vracející se erro převést na pq Error, aby server nevracel status 500
		// což je chyba na straně serveru, jelikož tento error se vrací pokud není danný uživatel v databázi,
		// což je chyba na straně klienta, takže by se mělo vracet spíš něco jako 403 (StatusForbidden)
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "foreign_key_violation", "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		// pokud chyba není způsobena neexistujícím uživatelem, ke kterému mají účty povinný vztah,
		// ani snaha vytvořit účet uživateli s měnou, který už existuje vrátí se Status
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
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

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	// The logged-in user can create account just him/her self so it is neccessary to
	// check if request for creating new acc is from same user as provided auth token.
	if account.Owner != authPayload.Username {
		err := errors.New("account doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
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
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	arg := db.ListAccountsParams{
		Owner:  authPayload.Username,
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
