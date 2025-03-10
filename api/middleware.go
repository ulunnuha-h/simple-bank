package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ulunnuha-h/simple_bank/token"
)

const (
	authHeaderKey = "authorization"
	authTypeBearer = "bearer"
	authPayloadKey = "auth_payload"
)

func AuthMiddleware(tokenGenerator token.Generator) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if len(ctx.GetHeader(authHeaderKey)) == 0 {
			err := errors.New("authorization header is not provided")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		fields := strings.Fields(ctx.GetHeader(authHeaderKey))
		if len(fields) < 2 {
			err := errors.New("invalid authorization header format")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		authType := strings.ToLower(fields[0])
		if authType != authTypeBearer {
			err := fmt.Errorf("unsupported authorization type %s", authType)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		accessToken := fields[1]
		payload, err := tokenGenerator.Verify(accessToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		ctx.Set(authPayloadKey, payload)
		ctx.Next()
	}
}

func GetAuthPayload(ctx *gin.Context) (*token.Payload, error) {
	payload, ok := ctx.Get(authPayloadKey)
	if !ok {
		return nil, token.ErrAuthNotProvided
	}

	authPayload, ok := payload.(*token.Payload)
	if !ok {
		return nil, token.ErrAuthNotProvided
	}

	return authPayload, nil
}