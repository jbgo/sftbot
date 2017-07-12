package trading

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"testing"

	"github.com/jbgo/sftbot/plx"
	"net/http"
	"path/filepath"
	"runtime"
)

func credentialsPath() string {
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)
	return filepath.Join(dir, "..", ".creds.json")
}

func TestGetMarket(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"BTC_ABC":{}}`)
	}))

	defer testServer.Close()

	fmt.Println(credentialsPath())
	client := plx.NewClient(testServer.URL, credentialsPath())
	exchange := NewPlxExchange(client)

	t.Run("existing market", func(t *testing.T) {
		market, err := exchange.GetMarket("BTC_ABC")
		require.Nil(t, err)
		assert.Implements(t, (*Market)(nil), market)
		assert.Equal(t, "ABC", market.GetCurrency())
	})

	t.Run("unknown market", func(t *testing.T) {
		market, err := exchange.GetMarket("BTC_DNE")
		require.Nil(t, market)
		require.NotNil(t, err)
		require.Contains(t, err.Error(), "unknown market")
	})
}

func TestGetBalance(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{
      "ABC": {
        "available":"0.42",
        "onOrders": "1.23",
        "btcValue": "0.29"
      }
    }`)
	}))

	defer testServer.Close()

	client := plx.NewClient(testServer.URL, credentialsPath())
	exchange := NewPlxExchange(client)

	t.Run("return balance", func(t *testing.T) {
		balance, err := exchange.GetBalance("ABC")
		require.Nil(t, err)
		require.NotNil(t, balance)
		assert.Equal(t, 0.42, balance.Available)
		assert.Equal(t, 1.23, balance.OnOrders)
		assert.Equal(t, 0.29, balance.BtcValue)
	})
}
