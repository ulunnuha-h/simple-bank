package api

import (
	"log"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ulunnuha-h/simple_bank/util"
)

func TestMain(m *testing.M) {
	err := util.LoadConfig("../.")
	if err != nil {
		log.Fatal("Failed to load .env file")
	}

	gin.SetMode(gin.TestMode)

	os.Exit(m.Run())
}
