package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	db "github.com/ulunnuha-h/simple_bank/db/sqlc"
	"github.com/ulunnuha-h/simple_bank/token"
)

type Server struct{
	store db.Store
	router *gin.Engine
	tokenGenerator token.Generator
}

func NewServer(store db.Store) (*Server, error){
	tokenGenerator, err := token.NewPasetoGenerator(viper.GetString("SECRET_KEY"))
	if err != nil {
		return nil, fmt.Errorf("cannot create token generator: %w", err)
	}

	server := &Server{
		store: store,
		tokenGenerator: tokenGenerator,
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", currencyValidator)
	}

	server.router = setupRouter(server)
	return server, nil
}

func setupRouter(server *Server) *gin.Engine{
	router := gin.Default()

	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)

	router.Use(AuthMiddleware(server.tokenGenerator))

	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.listAccount)
	router.DELETE("/accounts/:id", server.deleteAccount)
	router.PUT("/accounts/:id", server.updateAccount)

	router.POST("/transfers", server.createTransfer)
	return router
}

func (server *Server) Start(address string) error{
	return server.router.Run(address)
}

func errorResponse(err error) gin.H{	
	return gin.H{"error": err.Error()}
}