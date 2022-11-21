package api

import (
	"github.com/gin-gonic/gin"
	db "github.com/karlib/simple_bank/db/sqlc"
)

// Server servers all http requests for my bank service
type Server struct {
	store  *db.Store
	router *gin.Engine
}

// NewServer creates a new HTTP server instance and setup routing
func NewServer(store *db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	router.POST("/accounts", server.createAccount)

	//Add routes to router
	server.router = router
	return server
}
