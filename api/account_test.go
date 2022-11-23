package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	mockdb "github.com/karlib/simple_bank/db/mock"
	db "github.com/karlib/simple_bank/db/sqlc"
	"github.com/karlib/simple_bank/util"
	"github.com/stretchr/testify/require"
)

// Table driven test set to cover all scenarion on 100%

func TestGetAccountAPI(t *testing.T) {
	account := randomAccount()

	testCases := []struct {
		// struct deffinition for each case
		name          string
		accountID     int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		// different tested cases
		{
			name:      "OK",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				// build stubs
				// it means i expect that the GetAccount function of the store to be called with any context
				// and this specific account ID arguments
				// i can also specify how many times this function should be called
				// at the end we will define expected result with use Return(account, nil)
				// because we are expecting one specific account from db and empty error
				// arguments have to be equal to return type of GetAccount function inside the
				// Queier interface
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check response status code
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name:      "NotFound",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				// build stubs
				// it means i expect that the GetAccount function of the store to be called with any context
				// and this specific account ID arguments
				// i can also specify how many times this function should be called
				// at the end we will define expected result with use Return(account, nil)
				// because we are expecting one specific account from db and empty error
				// arguments have to be equal to return type of GetAccount function inside the
				// Queier interface
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check response status code
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				// build stubs
				// it means i expect that the GetAccount function of the store to be called with any context
				// and this specific account ID arguments
				// i can also specify how many times this function should be called
				// at the end we will define expected result with use Return(account, nil)
				// because we are expecting one specific account from db and empty error
				// arguments have to be equal to return type of GetAccount function inside the
				// Queier interface
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check response status code
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "InvalidID",
			accountID: 0,
			buildStubs: func(store *mockdb.MockStore) {
				// build stubs
				// it means i expect that the GetAccount function of the store to be called with any context
				// and this specific account ID arguments
				// i can also specify how many times this function should be called
				// at the end we will define expected result with use Return(account, nil)
				// because we are expecting one specific account from db and empty error
				// arguments have to be equal to return type of GetAccount function inside the
				// Queier interface
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check response status code
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		// TODO: add more cases
	}

	for i := range testCases {
		tc := testCases[i]
		// The each case will be run as a separate sub-test
		t.Run(tc.name, func(t *testing.T) {
			// end of run subtests
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)
			// start test server and send request
			server := newTestServer(t, store)

			// It creates new response writter
			// so we have not to start real listening server on our machine
			recorder := httptest.NewRecorder()
			// define url for testing request
			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			// testing api request
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)
			// This will send our Api request through the server router and send his response through recorder
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})

	}

}

func randomAccount() db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    util.RandomOwner(),
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotAccount db.Account
	err = json.Unmarshal(data, &gotAccount)
	require.NoError(t, err)
	require.Equal(t, account, gotAccount)
}
