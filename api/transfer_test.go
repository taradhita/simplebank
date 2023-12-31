package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	mockdb "github.com/taradhita/simplebank/db/mock"
	db "github.com/taradhita/simplebank/db/sqlc"
	"github.com/taradhita/simplebank/util"
)

func TestCreateTransferAPI(t *testing.T) {
	firstAccount := randomAccount()
	secondAccount := randomAccount()
	thirdAccount := randomAccount()

	firstAccount.Currency = util.USD
	secondAccount.Currency = util.USD
	thirdAccount.Currency = util.CAD

	amount := util.RandomInt(1, 100)

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"from_account_id": firstAccount.ID,
				"to_account_id":   secondAccount.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(firstAccount.ID)).Times(1).Return(firstAccount, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(secondAccount.ID)).Times(1).Return(secondAccount, nil)

				arg := db.TransferTxParams{
					FromAccountID: firstAccount.ID,
					ToAccountID:   secondAccount.ID,
					Amount:        amount,
				}

				store.EXPECT().TransferTx(gomock.Any(), gomock.Eq(arg)).Times(1)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"from_account_id": firstAccount.ID,
				"to_account_id":   secondAccount.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(firstAccount.ID)).Times(1).Return(firstAccount, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(secondAccount.ID)).Times(1).Return(secondAccount, nil)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(1).Return(db.TransferTxResult{}, sql.ErrTxDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "AccountNotFound",
			body: gin.H{
				"from_account_id": firstAccount.ID,
				"to_account_id":   secondAccount.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(firstAccount.ID)).Times(1).Return(firstAccount, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(secondAccount.ID)).Times(1).Return(db.Account{}, sql.ErrNoRows)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "ToCurrencyMismatch",
			body: gin.H{
				"from_account_id": firstAccount.ID,
				"to_account_id":   thirdAccount.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(firstAccount.ID)).Times(1).Return(firstAccount, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(thirdAccount.ID)).Times(1).Return(thirdAccount, nil)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},

		{
			name: "FromCurrencyMismatch",
			body: gin.H{
				"from_account_id": firstAccount.ID,
				"to_account_id":   thirdAccount.ID,
				"amount":          amount,
				"currency":        util.CAD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(firstAccount.ID)).Times(1).Return(firstAccount, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(thirdAccount.ID)).Times(0)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "NegativeAmount",
			body: gin.H{
				"from_account_id": firstAccount.ID,
				"to_account_id":   secondAccount.ID,
				"amount":          -amount,
				"currency":        util.USD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "GetAccountError",
			body: gin.H{
				"from_account_id": firstAccount.ID,
				"to_account_id":   secondAccount.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(1).Return(db.Account{}, sql.ErrConnDone)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			// start test server and send request
			server := NewServer(store)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, "/transfers", bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}
