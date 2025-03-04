package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	mockdb "github.com/ulunnuha-h/simple_bank/db/mock"
	db "github.com/ulunnuha-h/simple_bank/db/sqlc"
	"github.com/ulunnuha-h/simple_bank/util"
	"go.uber.org/mock/gomock"
)

func TestGetAccountAPI(t *testing.T){
	account := randomAccount();

	testCases := []struct{
		name string
		accountId int64
		buildStubs func(store *mockdb.MockStore)
		checkReposne func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			accountId: account.ID,
			buildStubs: func (store *mockdb.MockStore)  {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			checkReposne: func (t *testing.T, recorder *httptest.ResponseRecorder)  {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name: "NotFound",
			accountId: account.ID,
			buildStubs: func (store *mockdb.MockStore)  {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
			},
			checkReposne: func (t *testing.T, recorder *httptest.ResponseRecorder)  {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "InternalError",
			accountId: account.ID,
			buildStubs: func (store *mockdb.MockStore)  {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkReposne: func (t *testing.T, recorder *httptest.ResponseRecorder)  {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidID",
			accountId: 0,
			buildStubs: func (store *mockdb.MockStore)  {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkReposne: func (t *testing.T, recorder *httptest.ResponseRecorder)  {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases{
		tc := testCases[i]
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		store := mockdb.NewMockStore(ctrl)
		tc.buildStubs(store)

		server := NewServer(store)
		recorder := httptest.NewRecorder()

		url := fmt.Sprintf("/accounts/%d", tc.accountId)
		request, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)

		server.router.ServeHTTP(recorder, request)
		tc.checkReposne(t, recorder)
	}

}

func randomAccount() db.Account {
	return db.Account{
		ID: util.RandomInt(1, 1000),
		Owner: util.RandomOwner(),
		Balance: util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account){
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotAccount db.Account
	err = json.Unmarshal(data, &gotAccount)
	require.NoError(t, err)
	require.Equal(t, account, gotAccount)
}