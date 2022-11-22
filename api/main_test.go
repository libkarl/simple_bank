package api

import (
	"github.com/gin-gonic/gin"
	"os"
	"testing"
)

// it is used to set gin for using testmode insted of debug
// this step will decrease the mess in logs in the process of package testing

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}