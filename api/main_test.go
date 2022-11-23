package api

import (
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/karlib/simple_bank/db/sqlc"
	"github.com/karlib/simple_bank/util"
	"github.com/stretchr/testify/require"
)

// it is used to set gin for using testmode insted of debug
// this step will decrease the mess in logs in the process of package testing

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

// function which will create new server for test 

func newTestServer(t *testing.T, store db.Store ) *Server{
	config := util.Config{
		TokenSymetricKey: util.RandomString(32),
		AccessTokenDuration: time.Minute,
	}


	server, err := NewServer(config, store)
	require.NoError(t, err)

	return server
}