package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	db "github.com/ulunnuha-h/simple_bank/db/sqlc"
	"github.com/ulunnuha-h/simple_bank/util"
)

type createUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

type UserReponse struct{
	Username          string    `json:"username"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}

func newUserReponse(user db.User) UserReponse {
	return UserReponse{
		Username: user.Username,
		FullName: user.FullName,
		Email: user.Email,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt: user.CreatedAt,
	}
}

func (server *Server) createUser(ctx *gin.Context){
	var	req createUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil{
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	args := db.CreateUserParams{
		Username: req.Username,
		HashedPassword: hashedPassword,
		FullName: req.FullName,
		Email: req.Email,
	}

	user, err := server.store.CreateUser(ctx, args)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return	
			}
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	createUserReponse := newUserReponse(user)

	ctx.JSON(http.StatusOK, createUserReponse)
}

type loginUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`	
}

type loginUserReponse struct {
	AccessToken string `json:"access_token"`
	LoggedUser UserReponse `json:"logged_user"`
}

func (server *Server) loginUser(ctx *gin.Context){
	var req loginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil{
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUser(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	err = util.CheckPassword(req.Password, user.HashedPassword)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	_, accessToken, err := server.tokenGenerator.CreateToken(req.Username, 10 * time.Minute)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	refreshPayload, refreshToken, err := server.tokenGenerator.CreateToken(req.Username, time.Hour)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	args := db.CreateSessionParams{
		ID: string(refreshPayload.ID.String()),
		Username: req.Username,
		RefreshToken: refreshToken,
		UserAgent: ctx.Request.UserAgent(),
		ClientIp: ctx.ClientIP(),
		IsBlocked: false,
		ExpiredAt: time.Now().Add(time.Hour),
	}

	server.store.CreateSession(ctx, args)

	ctx.SetCookie("refresh_token",refreshToken, int(time.Hour), "/", "localhost", false, true)

	ctx.JSON(http.StatusOK, loginUserReponse{
		AccessToken: accessToken,
		LoggedUser: newUserReponse(user),
	})
}

type RefreshTokenResponse struct{
	Username    string    `json:"username"`
	AccessToken string `json:"access_token"`
}

func (server *Server) refreshToken(ctx *gin.Context){
	token, err := ctx.Cookie("refresh_token")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return	
	}

	payload, err := server.tokenGenerator.Verify(token)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	session, err := server.store.GetSession(ctx, payload.ID.String())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if session.IsBlocked || time.Now().After(session.ExpiredAt) {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	_, accessToken, err := server.tokenGenerator.CreateToken(session.Username, 10 * time.Minute)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, RefreshTokenResponse{
		Username: session.Username,
		AccessToken: accessToken,
	})
}