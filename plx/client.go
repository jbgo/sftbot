package plx

const LIVE_URL = "https://poloniex.com"

type Client struct {
	BaseUrl string
}

func (client *Client) PublicApiUrl() string {
	return client.BaseUrl + "/public"
}

func (client *Client) TradingApiUrl() string {
	return client.BaseUrl + "/tradingApi"
}
