package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	db "github.com/karlib/simple_bank/db/sqlc"
	"github.com/karlib/simple_bank/token"
	"github.com/karlib/simple_bank/util"
)

// Server servers all http requests for my bank service
type Server struct {
	config     util.Config
	store      db.Store
	router     *gin.Engine
	tokenMaker token.Maker
}

// NewServer creates a new HTTP server instance and setup routing
func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPassetoMaker(config.TokenSymetricKey)
	if err != nil {
		// %w is used to wrap original error
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}
	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}

	// Here i got access to the actual used validate engine for GIN framework
	// , at the end with .(*validator.Validate) will convert the output to the type *validator.Validate
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// The first argument is the name of the validation tag, the second argument is
		// custom custom validation function wihich will be registe for this specific tag
		// so if for struct field inside request will be used tag like this-> binding:"currency"
		// then the GIN framework will use my custom currency validator
		v.RegisterValidation("currency", validCurrency)
	}

	server.setupRouter()

	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()
	// all registred routes
	// the the first two routes must be public the rest of will be protected by authMiddleware

	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)

	authRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker))
	authRoutes.POST("/accounts", server.createAccount)
	authRoutes.GET("/accounts/:id", server.getAccountByID)
	authRoutes.GET("/accounts", server.listAccount)
	authRoutes.POST("/transfers", server.createTransfer)
	//Add routes to router
	server.router = router
}

// Start runs HTTP server on a specific address
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
