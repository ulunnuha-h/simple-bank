package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/ulunnuha-h/simple_bank/token"
)

func addAuthorization(
	t *testing.T,
	request *http.Request,
	tokenGenerator token.Generator,
	authType string,
	username string,
	duration time.Duration,
){
	_, token, err := tokenGenerator.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	request.Header.Set(authHeaderKey, fmt.Sprintf("%s %s", authType, token))
}

func TestAuthMiddleware(t *testing.T){
	testcases := []struct{
		name string
		setupAuth func(t *testing.T, request *http.Request, tokenGenerator token.Generator)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, request *http.Request, tokenGenerator token.Generator){
				addAuthorization(t, request, tokenGenerator, authTypeBearer, "user", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder){
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "NoAuthHeader",
			setupAuth: func(t *testing.T, request *http.Request, tokenGenerator token.Generator){},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder){
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "ExpiredToken",
			setupAuth: func(t *testing.T, request *http.Request, tokenGenerator token.Generator){
				addAuthorization(t, request, tokenGenerator, authTypeBearer, "user", -time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder){
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for i := range testcases {
		tc := testcases[i]

		t.Run(tc.name, func(t *testing.T) {
			server, _ := NewServer(nil)

			authPath := "/auth"
			server.router.GET(
				authPath, 
				AuthMiddleware(server.tokenGenerator), 
				func(ctx *gin.Context) {
					ctx.JSON(http.StatusOK, gin.H{})
				})

			recorder := httptest.NewRecorder()
			request, err := http.NewRequest(http.MethodGet, authPath, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenGenerator)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}