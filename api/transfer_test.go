package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	mockdb "github.com/ulunnuha-h/simple_bank/db/mock"
	db "github.com/ulunnuha-h/simple_bank/db/sqlc"
	"go.uber.org/mock/gomock"
)

func TestCreateTransferAPI(t *testing.T){
	fromAccount := randomAccount();
	toAccount := randomAccount();
	fromAccount.Currency = "IDR"
	toAccount.Currency = "IDR"

	transferAmount := fromAccount.Balance/2
	transferResult := db.TransferTxResult{
		Transfer: db.Transfer{
			FromAccountID: fromAccount.ID,
			ToAccountID: toAccount.ID,
			Amount: transferAmount,
		},
		FromAccount: fromAccount,
		ToAccount: toAccount,
		FromEntry: db.Entry{
			AccountID: fromAccount.ID,
			Amount: -transferAmount,
		},
		ToEntry: db.Entry{
			AccountID: toAccount.ID,
			Amount: transferAmount,
		},
	}

	testCases := []struct{
		name string
		requestBody createTransferRequest
		buildStubs func(store *mockdb.MockStore)
		checkReposne func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			requestBody: createTransferRequest{
				FromAccountId: fromAccount.ID,
				ToAccountId: toAccount.ID,
				Amount: fromAccount.Balance/2,
				Currency: "IDR",
			},
			buildStubs: func (store *mockdb.MockStore)  {
				args := db.TransferTxParams{	
					FromAccountId: fromAccount.ID,
					ToAccountId: toAccount.ID,
					Amount: transferAmount,
				}

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).
					Times(1).
					Return(fromAccount, nil)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).
					Times(1).
					Return(toAccount, nil)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Eq(args)).
					Times(1).
					Return(transferResult, nil)
			},
			checkReposne: func (t *testing.T, recorder *httptest.ResponseRecorder)  {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchTransfer(t, recorder.Body, transferResult)
			},
		},
		{
			name: "InternalError",
			requestBody: createTransferRequest{
				FromAccountId: fromAccount.ID,
				ToAccountId: toAccount.ID,
				Amount: fromAccount.Balance/2,
				Currency: "IDR",
			},
			buildStubs: func (store *mockdb.MockStore)  {
				args := db.TransferTxParams{	
					FromAccountId: fromAccount.ID,
					ToAccountId: toAccount.ID,
					Amount: transferAmount,
				}

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).
					Times(1).
					Return(fromAccount, nil)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).
					Times(1).
					Return(toAccount, nil)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Eq(args)).
					Times(1).
					Return(db.TransferTxResult{}, sql.ErrConnDone)

			},
			checkReposne: func (t *testing.T, recorder *httptest.ResponseRecorder)  {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "EmptyBody",
			requestBody: createTransferRequest{},
			buildStubs: func (store *mockdb.MockStore)  {
				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkReposne: func (t *testing.T, recorder *httptest.ResponseRecorder)  {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "MismatchCurrency",
			requestBody: createTransferRequest{
				FromAccountId: fromAccount.ID,
				ToAccountId: toAccount.ID,
				Amount: fromAccount.Balance/2,
				Currency: "EUR",
			},
			buildStubs: func (store *mockdb.MockStore)  {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).
					Times(1).
					Return(fromAccount, nil)

			},
			checkReposne: func (t *testing.T, recorder *httptest.ResponseRecorder)  {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InsufficientBalance",
			requestBody: createTransferRequest{
				FromAccountId: fromAccount.ID,
				ToAccountId: toAccount.ID,
				Amount: fromAccount.Balance + 1,
				Currency: "IDR",
			},
			buildStubs: func (store *mockdb.MockStore)  {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).
					Times(1).
					Return(fromAccount, nil)
			},
			checkReposne: func (t *testing.T, recorder *httptest.ResponseRecorder)  {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "NotFound",
			requestBody: createTransferRequest{
				FromAccountId: fromAccount.ID + toAccount.ID,
				ToAccountId: toAccount.ID,
				Amount: fromAccount.Balance,
				Currency: "IDR",
			},
			buildStubs: func (store *mockdb.MockStore)  {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID + toAccount.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
			},
			checkReposne: func (t *testing.T, recorder *httptest.ResponseRecorder)  {
				require.Equal(t, http.StatusNotFound, recorder.Code)
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

		url := "/transfers"
		jsonData, err := json.Marshal(tc.requestBody)
		require.NoError(t, err)

		request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
		require.NoError(t, err)

		server.router.ServeHTTP(recorder, request)
		tc.checkReposne(t, recorder)
	}

}

func requireBodyMatchTransfer(t *testing.T, body *bytes.Buffer, account db.TransferTxResult){
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotResult db.TransferTxResult
	err = json.Unmarshal(data, &gotResult)
	require.NoError(t, err)
	require.Equal(t, account, gotResult)
}