package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	db "github.com/karlib/simple_bank/db/sqlc"
)

// Server servers all http requests for my bank service
type Server struct {
	store  db.Store
	router *gin.Engine
}

// NewServer creates a new HTTP server instance and setup routing
func NewServer(store db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()
	// Here i got access to the actual used validate engine for GIN framework
	// , at the end with .(*validator.Validate) will convert the output to the type *validator.Validate 
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// The first argument is the name of the validation tag, the second argument is 
		// custom custom validation function wihich will be registe for this specific tag
		// so if for struct field inside request will be used tag like this-> binding:"currency"
		// then the GIN framework will use my custom currency validator
		v.RegisterValidation("currency", validCurrency)
	}

	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccountByID)
	router.GET("/accounts", server.listAccount)

	router.POST("/transfers", server.createTransfer)
	//Add routes to router
	server.router = router
	return server
}

// Start runs HTTP server on a specific address
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
