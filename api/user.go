package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/karlib/simple_bank/db/sqlc"
	"github.com/karlib/simple_bank/util"
	"github.com/lib/pq"
)

// alpanum tag is use for ban all special characters inside the username

type createUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

type userResponse struct {
	Username          string    `json:"username"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}

// This fn will convert user db.User to userResponse (after authentication)
func newUserResponse(user db.User) userResponse{
	return userResponse{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}
}

func (server *Server) createUser(ctx *gin.Context) {
	var req createUserRequest
	// pokud error není nil klient poskytl nesprávné údaje
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	hashedPassword, err := util.HashPassword(req.Password)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashedPassword,
		FullName:       req.FullName,
		Email:          req.Email,
	}

	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		// tento email se vrací pokud stejný uživatel už existuje, v takovém případě je třeba, aby se typ erroru převedl na typ chyby kterou
		// vrací databáze se status codem 403
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		// pokud se jedná o jinou chybu než je duplicita v databází vrací iternall error ( code 500)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	response := newUserResponse(user)

	ctx.JSON(http.StatusOK, response)
}

type loginUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginUserResponse struct {
	AccessToken string `json:"access_token`
	User userResponse `json:"user"`
}

func (server *Server) loginUser(ctx *gin.Context) {
	var req loginUserRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return 
	}

	// if there is not an error we will find user in the database with function GetUser()
 
	user, err := server.store.GetUser(ctx, req.Username)
	if err !=nil {
		// pokud je tady error může to být ze dvou důvodu 1. user neexistuje takže ErrNoRows
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse((err)))
			return 
		}
		// nejaký nečekaný error při kontaktu s databází (není dostupná)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return 
	}

	// The folowing code will check if the password provided by client is valid or not.
	err = util.CheckPassword(req.Password, user.HashedPassword)
	// If this func returns error, it means the provided password is incorrect
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return 
	}

	// So now i already know that the provided password is correct so i can screate assessToken for user
	accsessToken, err := server.tokenMaker.CreateToken(
		user.Username,
		server.config.AccessTokenDuration,
	)
	// If the unexpected error occurs than the server will send response with StatusInternalServerError
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
	}


	rsp := loginUserResponse{
		AccessToken: accsessToken,
		User: newUserResponse(user),
	}

	// response is sended to the client
	ctx.JSON(http.StatusOK, rsp)
}