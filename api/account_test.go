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

		server, err := NewServer(store)
		require.NoError(t, err)
		recorder := httptest.NewRecorder()

		url := fmt.Sprintf("/accounts/%d", tc.accountId)
		request, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)

		server.router.ServeHTTP(recorder, request)
		tc.checkReposne(t, recorder)
	}

}

func TestCreateAccountAPI(t *testing.T){
	testAccount := randomAccount();
	args := db.CreateAccountParams{
		Owner: testAccount.Owner,
		Balance: 0,
		Currency: testAccount.Currency,
	}

	testCases := []struct{
		name string
		account db.Account
		buildStubs func(store *mockdb.MockStore)
		checkReposne func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			account: testAccount,
			buildStubs: func (store *mockdb.MockStore)  {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(args)).
					Times(1).
					Return(testAccount, nil)
			},
			checkReposne: func (t *testing.T, recorder *httptest.ResponseRecorder)  {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, testAccount)
			},
		},
		{
			name: "InternalError",
			account: testAccount,
			buildStubs: func (store *mockdb.MockStore)  {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(args)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkReposne: func (t *testing.T, recorder *httptest.ResponseRecorder)  {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "EmptyBody",
			account: db.Account{},
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

		server, err := NewServer(store)
		require.NoError(t, err)
		recorder := httptest.NewRecorder()

		url := "/accounts"
		accountRequest := &createAccountRequest{
			Currency: tc.account.Currency,
		}

		jsonData, err := json.Marshal(accountRequest)
		require.NoError(t, err)

		request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
		require.NoError(t, err)

		server.router.ServeHTTP(recorder, request)
		tc.checkReposne(t, recorder)
	}
}

func TestListAccountAPI(t *testing.T){
	var n = 5
	accounts := make([]db.Account, n)
	for i := range n {
		accounts[i] = randomAccount()
	}

	testCases := []struct{
		name string
		query listAccountRequest
		buildStubs func(store *mockdb.MockStore)
		checkReposne func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			query: listAccountRequest{
				PAGE_ID: 1,
				PAGE_SIZE: 5,
			},
			buildStubs: func (store *mockdb.MockStore)  {
				args := db.ListAccountsParams{
					Limit: 5,
					Offset: 0,
				}

				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Eq(args)).
					Times(1).
					Return(accounts, nil)
			},
			checkReposne: func (t *testing.T, recorder *httptest.ResponseRecorder)  {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccounts(t, recorder.Body, accounts)
			},
		},
		{
			name: "InternalError",
			query: listAccountRequest{
				PAGE_ID: 1,
				PAGE_SIZE: 5,
			},
			buildStubs: func (store *mockdb.MockStore)  {
				args := db.ListAccountsParams{
					Limit: 5,
					Offset: 0,
				}

				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Eq(args)).
					Times(1).
					Return([]db.Account{}, sql.ErrConnDone)
			},
			checkReposne: func (t *testing.T, recorder *httptest.ResponseRecorder)  {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidQuery",
			query: listAccountRequest{
				PAGE_ID: 0,
				PAGE_SIZE: 4,
			},
			buildStubs: func (store *mockdb.MockStore)  {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
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

		server, err := NewServer(store)
		require.NoError(t, err)
		recorder := httptest.NewRecorder()

		url := fmt.Sprintf("/accounts?page_id=%d&page_size=%d", tc.query.PAGE_ID, tc.query.PAGE_SIZE) 
		request, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)

		server.router.ServeHTTP(recorder, request)
		tc.checkReposne(t, recorder)
	}
}

func TestDeleteAccountAPI(t *testing.T){
	account := randomAccount()

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

				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(nil)
			},
			checkReposne: func (t *testing.T, recorder *httptest.ResponseRecorder)  {
				require.Equal(t, http.StatusOK, recorder.Code)
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

				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkReposne: func (t *testing.T, recorder *httptest.ResponseRecorder)  {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "InternalErrorOnGet",
			accountId: account.ID,
			buildStubs: func (store *mockdb.MockStore)  {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)

				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkReposne: func (t *testing.T, recorder *httptest.ResponseRecorder)  {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InternalErrorOnDelete",
			accountId: account.ID,
			buildStubs: func (store *mockdb.MockStore)  {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)

				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(sql.ErrConnDone)
			},
			checkReposne: func (t *testing.T, recorder *httptest.ResponseRecorder)  {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "BadRequest",
			accountId: 0,
			buildStubs: func (store *mockdb.MockStore)  {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Any()).
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

		server, err := NewServer(store)
		require.NoError(t, err)
		recorder := httptest.NewRecorder()

		url := fmt.Sprintf("/accounts/%d", tc.accountId) 
		request, err := http.NewRequest(http.MethodDelete, url, nil)
		require.NoError(t, err)

		server.router.ServeHTTP(recorder, request)
		tc.checkReposne(t, recorder)
	}
}

func TestUpdateAccountAPI(t *testing.T){
	account := randomAccount()
	updatedAccount := db.Account{
		ID: account.ID,
		Owner: account.Owner,
		Balance: util.RandomMoney(),
		Currency: account.Currency,
		CreatedAt: account.CreatedAt,
	}

	testCases := []struct{
		name string
		accountId int64
		requestBody updateAccountJsonRequest
		buildStubs func(store *mockdb.MockStore)
		checkReposne func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			accountId: account.ID,
			requestBody: updateAccountJsonRequest{
				Balance: updatedAccount.Balance,
			},
			buildStubs: func (store *mockdb.MockStore)  {
				args := db.UpdateAccountParams{
					ID: account.ID,
					Balance: updatedAccount.Balance,
				}

				store.EXPECT().
					UpdateAccount(gomock.Any(), gomock.Eq(args)).
					Times(1).
					Return(updatedAccount, nil)
			},
			checkReposne: func (t *testing.T, recorder *httptest.ResponseRecorder)  {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, updatedAccount)
			},
		},
		{
			name: "InternalError",
			accountId: account.ID,
			requestBody: updateAccountJsonRequest{
				Balance: updatedAccount.Balance,
			},
			buildStubs: func (store *mockdb.MockStore)  {
				args := db.UpdateAccountParams{
					ID: account.ID,
					Balance: updatedAccount.Balance,
				}

				store.EXPECT().
					UpdateAccount(gomock.Any(), gomock.Eq(args)).
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
			requestBody: updateAccountJsonRequest{
				Balance: updatedAccount.Balance,
			},
			buildStubs: func (store *mockdb.MockStore)  {
				store.EXPECT().
					UpdateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkReposne: func (t *testing.T, recorder *httptest.ResponseRecorder)  {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "EmptyBalance",
			accountId: account.ID,
			requestBody: updateAccountJsonRequest{},
			buildStubs: func (store *mockdb.MockStore)  {
				store.EXPECT().
					UpdateAccount(gomock.Any(), gomock.Any()).
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

		server, err := NewServer(store)
		require.NoError(t, err)
		recorder := httptest.NewRecorder()

		url := fmt.Sprintf("/accounts/%d", tc.accountId) 
		jsonData, err := json.Marshal(tc.requestBody)
		require.NoError(t, err)

		request, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonData))
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

func requireBodyMatchAccounts(t *testing.T, body *bytes.Buffer, account []db.Account){
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotAccount []db.Account
	err = json.Unmarshal(data, &gotAccount)
	require.NoError(t, err)
	require.Equal(t, account, gotAccount)
}