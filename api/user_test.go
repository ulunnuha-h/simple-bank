package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	mockdb "github.com/ulunnuha-h/simple_bank/db/mock"
	db "github.com/ulunnuha-h/simple_bank/db/sqlc"
	"github.com/ulunnuha-h/simple_bank/util"
	"go.uber.org/mock/gomock"
)

type passwordMatcher struct {
	arg db.CreateUserParams
	password string
}

func (e passwordMatcher) Matches(x any) bool {
	arg, ok := x.(db.CreateUserParams)
	if !ok {
		return false
	}

	err := util.CheckPassword(e.password, arg.HashedPassword)
	if err != nil {
		return false
	}

	e.arg.HashedPassword = arg.HashedPassword
	return reflect.DeepEqual(e.arg, arg)
}

func (e passwordMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)
}

func TestCreateUserAPI(t *testing.T){
	testUser, password := randomUser()

	testCases := []struct{
		name string
		requestBody createUserRequest
		buildStubs func(store *mockdb.MockStore)
		checkReposne func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			requestBody: createUserRequest{
				Username: testUser.Username,
				FullName: testUser.FullName,
				Email: testUser.Email,
				Password: password,
			},
			buildStubs: func (store *mockdb.MockStore)  {
				arg := db.CreateUserParams{
					Username: testUser.Username,
					FullName: testUser.FullName,
					HashedPassword: password,
					Email: testUser.Email,
				}

				store.EXPECT().
					CreateUser(gomock.Any(), passwordMatcher{arg, password}).
					Times(1).
					Return(testUser, nil)
			},
			checkReposne: func (t *testing.T, recorder *httptest.ResponseRecorder)  {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchUser(t, recorder.Body, testUser)
			},
		},
		{
			name: "InternalError",
			requestBody: createUserRequest{
				Username: testUser.Username,
				FullName: testUser.FullName,
				Email: testUser.Email,
				Password: password,
			},
			buildStubs: func (store *mockdb.MockStore)  {
				arg := db.CreateUserParams{
					Username: testUser.Username,
					FullName: testUser.FullName,
					HashedPassword: password,
					Email: testUser.Email,
				}

				store.EXPECT().
					CreateUser(gomock.Any(), passwordMatcher{arg, password}).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
			},
			checkReposne: func (t *testing.T, recorder *httptest.ResponseRecorder)  {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "DuplicatedUsername",
			requestBody: createUserRequest{
				Username: testUser.Username,
				FullName: testUser.FullName,
				Email: testUser.Email,
				Password: password,
			},
			buildStubs: func (store *mockdb.MockStore)  {
				arg := db.CreateUserParams{
					Username: testUser.Username,
					FullName: testUser.FullName,
					HashedPassword: password,
					Email: testUser.Email,
				}

				store.EXPECT().
					CreateUser(gomock.Any(), passwordMatcher{arg, password}).
					Times(1).
					Return(db.User{}, &pq.Error{Code: "23505"})
			},
			checkReposne: func (t *testing.T, recorder *httptest.ResponseRecorder)  {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "InvalidUsername",
			requestBody: createUserRequest{
				Username: "username#1",
				FullName: testUser.FullName,
				Email: testUser.Email,
				Password: password,
			},
			buildStubs: func (store *mockdb.MockStore)  {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkReposne: func (t *testing.T, recorder *httptest.ResponseRecorder)  {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "TooLongPassword",
			requestBody: createUserRequest{
				Username: testUser.Username,
				FullName: testUser.FullName,
				Email: testUser.Email,
				Password: "nFXGPwKMcR9twfStvgg68Jf3r3jd6fevUzLFWP0DjU7H0c7699YPnQDCNrjAQj4DnDNjyYg8w",
			},
			buildStubs: func (store *mockdb.MockStore)  {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkReposne: func (t *testing.T, recorder *httptest.ResponseRecorder)  {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases{
		tc := testCases[i]
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		store := mockdb.NewMockStore(ctrl)
		tc.buildStubs(store)

		server, err := NewServer(store)
		require.NoError(t, err)
		recorder := httptest.NewRecorder()

		url := "/users"
		createUserRequest := tc.requestBody

		jsonData, err := json.Marshal(createUserRequest)
		require.NoError(t, err)

		request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
		require.NoError(t, err)

		server.router.ServeHTTP(recorder, request)
		tc.checkReposne(t, recorder)
	}
}

func randomUser() (db.User, string) {
	return db.User{
		Username: util.RandomString(6),
		FullName: util.RandomString(6),
		Email: util.RandomEmail(),
	}, util.RandomString(6)
}

func requireBodyMatchUser(t *testing.T, body *bytes.Buffer, user db.User){
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotUser db.User
	err = json.Unmarshal(data, &gotUser)
	require.NoError(t, err)
	require.Equal(t, user, gotUser)
}