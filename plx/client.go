package plx

const LIVE_URL = "https://poloniex.com"
const LIVE_CREDENTIALS_PATH = "./.creds.json"

type Client struct {
	BaseUrl         string
	CredentialsPath string
}

func NewClient(baseUrl, credentialsPath string) *Client {
	return &Client{
		BaseUrl:         baseUrl,
		CredentialsPath: credentialsPath,
	}
}

func NewLiveClient() *Client {
	return NewClient(LIVE_URL, LIVE_CREDENTIALS_PATH)
}

func (client *Client) PublicApiUrl() string {
	return client.BaseUrl + "/public"
}

func (client *Client) TradingApiUrl() string {
	return client.BaseUrl + "/tradingApi"
}
