package trading

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"testing"

	"github.com/jbgo/sftbot/plx"
	"net/http"
)

type TestPlxApiRequest struct {
	Endpoint string
	Command  string
	Response string
}

func matchesPublicApiRequest(testReq *TestPlxApiRequest, r *http.Request) bool {
	return r.URL.Path == "/public" &&
		testReq.Endpoint == "/public" &&
		r.Method == "GET" &&
		r.URL.Query().Get("command") == testReq.Command
}

func matchesTradingApiRequest(testReq *TestPlxApiRequest, r *http.Request) bool {
	return r.URL.Path == "/tradingApi" &&
		testReq.Endpoint == "/tradingApi" &&
		r.Method == "POST" &&
		r.PostFormValue("command") == testReq.Command
}

func buildTestServer(requests []*TestPlxApiRequest) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var matchedReq *TestPlxApiRequest

		for _, testReq := range requests {
			if testReq.Endpoint != r.URL.Path {
				continue
			}

			if matchesPublicApiRequest(testReq, r) || matchesTradingApiRequest(testReq, r) {
				matchedReq = testReq
				break
			}
		}

		if matchedReq != nil {
			fmt.Fprintln(w, matchedReq.Response)
		}
	}))
}

func TestMarket(t *testing.T) {
	testServer := buildTestServer([]*TestPlxApiRequest{
		&TestPlxApiRequest{"/public", "returnTicker", `{"BTC_ABC":{"last":"0.015"}}`},
		&TestPlxApiRequest{"/tradingApi", "returnOpenOrders", `[
      {"orderNumber":"123","type":"buy","rate":"4.56","amount":"12.0","total":"22.2"},
      {"orderNumber":"321","type":"sell","rate":"6.54","amount":"0.12","total":"0.0222"}
      ]`},
		&TestPlxApiRequest{"/public", "returnChartData", `[
      {"date":1405699200,"high":0.0045388,"low":0.00403001,"open":0.00404545,"close":0.00427592,"volume":44.11655644,"quoteVolume":10259.29079097,"weightedAverage":0.00430015}
      ]`},
		&TestPlxApiRequest{"/tradingApi", "buy", `{"orderNumber":123456,"resultingTrades":[{"amount":"42.0"}]}`},
		&TestPlxApiRequest{"/tradingApi", "sell", `{"orderNumber":654321,"resultingTrades":[{"amount":"0.376"}]}`},
	})

	defer testServer.Close()

	client := plx.NewClient(testServer.URL, credentialsPath())
	market, err := NewPlxMarket("BTC_ABC", client)
	require.Nil(t, err)

	t.Run("GetName", func(t *testing.T) {
		assert.Equal(t, "BTC_ABC", market.GetName())
	})

	t.Run("GetCurrency", func(t *testing.T) {
		assert.Equal(t, "ABC", market.GetCurrency())
	})

	t.Run("GetCurrentPrice", func(t *testing.T) {
		price, err := market.GetCurrentPrice()
		require.Nil(t, err)
		assert.Equal(t, 0.015, price)
	})

	t.Run("GetPendingOrders", func(t *testing.T) {
		orders, err := market.GetPendingOrders()
		require.Nil(t, err)
		require.Equal(t, 2, len(orders))

		o1 := orders[0]
		assert.Equal(t, "123", o1.Id)
		assert.Equal(t, "buy", o1.Type)
		assert.Equal(t, 4.56, o1.Price)
		assert.Equal(t, 12.0, o1.Amount)
		assert.Equal(t, 22.2, o1.Total)
		assert.Equal(t, false, o1.Filled)

		o2 := orders[1]
		assert.Equal(t, "321", o2.Id)
		assert.Equal(t, "sell", o2.Type)
		assert.Equal(t, 6.54, o2.Price)
		assert.Equal(t, 0.12, o2.Amount)
		assert.Equal(t, 0.0222, o2.Total)
		assert.Equal(t, false, o2.Filled)
	})

	t.Run("GetSummaryData", func(t *testing.T) {
		samples, err := market.GetSummaryData(1, 2)
		require.Nil(t, err)
		require.Equal(t, 1, len(samples))

		d := samples[0]
		assert.Equal(t, int64(1405699200), d.Date)
		assert.Equal(t, 0.0045388, d.High)
		assert.Equal(t, 0.00403001, d.Low)
		assert.Equal(t, 0.00404545, d.Open)
		assert.Equal(t, 0.00427592, d.Close)
		assert.Equal(t, 44.11655644, d.Volume)
		assert.Equal(t, 10259.29079097, d.QuoteVolume)
		assert.Equal(t, 0.00430015, d.WeightedAverage)
	})

	t.Run("Buy", func(t *testing.T) {
		order := &Order{}
		err := market.Buy(order)
		require.Nil(t, err)
		assert.Equal(t, "123456", order.Id)
	})

	t.Run("Sell", func(t *testing.T) {
		order := &Order{}
		err := market.Sell(order)
		require.Nil(t, err)
		assert.Equal(t, "654321", order.Id)
	})
}
