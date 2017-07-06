package plx

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const POLONIEX_TRADING_API_URL = "https://poloniex.com/tradingApi"
const API_CREDENTIALS_FILE = "./.creds.json"

type CompleteBalance struct {
	Currency  string
	Available float64 `json:"available"`
	OnOrders  float64 `json:"onOrders"`
	BtcValue  float64 `json:"btcValue"`
}

func GetBalance(currency string) (balance *CompleteBalance, err error) {
	balances, err := CompleteBalances()
	if err != nil {
		return balance, err
	}

	for _, b := range balances {
		if b.Currency == currency {
			balance = &b
		}
	}

	if balance == nil {
		err = fmt.Errorf("could not find balance for currency: %s", currency)
	}

	return balance, err
}

func CompleteBalances() ([]CompleteBalance, error) {
	balances := make([]CompleteBalance, 0)
	client := NewTradingApiClient()

	values := &url.Values{}
	values.Set("command", "returnCompleteBalances")

	resp, err := client.Post(values)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("request failed: %s\n\n", resp.Status, bytes.NewBuffer(body).String())
	}

	respData := make(map[string]map[string]string)

	err = json.Unmarshal(body, &respData)

	for key, data := range respData {
		balance := CompleteBalance{}
		balance.Currency = key
		balance.Available, _ = strconv.ParseFloat(data["available"], 64)
		balance.OnOrders, _ = strconv.ParseFloat(data["onOrders"], 64)
		balance.BtcValue, _ = strconv.ParseFloat(data["btcValue"], 64)
		balances = append(balances, balance)
	}

	return balances, nil
}

type OpenOrder struct {
	Number int64 `json:"orderNumber,string"`
	Type   string
	Rate   float64 `json:",string"`
	Amount float64 `json:",string"`
	Total  float64 `json:",string"`
}

func AllOpenOrders() (marketOrders map[string][]OpenOrder, err error) {
	marketOrders = make(map[string][]OpenOrder)
	client := NewTradingApiClient()

	values := &url.Values{}
	values.Set("command", "returnOpenOrders")
	values.Set("currencyPair", "all")

	resp, err := client.Post(values)

	if err != nil {
		return marketOrders, err
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return marketOrders, fmt.Errorf("request failed: %s\n\n", resp.Status, bytes.NewBuffer(body).String())
	}

	err = json.Unmarshal(body, &marketOrders)

	return marketOrders, err
}

type PlxPrivateTrade struct {
	GlobalTradeId int64
	TradeId       string
	OrderNumber   string
	Date          string
	Rate          float64 `json:",string"`
	Amount        float64 `json:",string"`
	Total         float64 `json:",string"`
	Fee           float64 `json:",string"`
	Type          string
	Category      string
}

// returnTradeHistory
// Returns your trade history for a given market, specified by the "currencyPair" POST parameter. You may specify "all" as the currencyPair to receive your trade history for all markets. You may optionally specify a range via "start" and/or "end" POST parameters, given in UNIX timestamp format; if you do not specify a range, it will be limited to one day. Sample output:
// [{ "globalTradeID": 25129732, "tradeID": "6325758", "date": "2016-04-05 08:08:40", "rate": "0.02565498", "amount": "0.10000000", "total": "0.00256549", "fee": "0.00200000", "orderNumber": "34225313575", "type": "sell", "category": "exchange" }, { "globalTradeID": 25129628, "tradeID": "6325741", "date": "2016-04-05 08:07:55", "rate": "0.02565499", "amount": "0.10000000", "total": "0.00256549", "fee": "0.00200000", "orderNumber": "34225195693", "type": "buy", "category": "exchange" }, ... ]

func MyTradeHistory(marketName string, startTime, endTime int64) (trades map[string][]*PlxPrivateTrade, err error) {
	client := NewTradingApiClient()

	values := &url.Values{}
	values.Set("command", "returnTradeHistory")
	values.Set("currencyPair", marketName)
	values.Set("start", strconv.FormatInt(startTime, 10))
	values.Set("end", strconv.FormatInt(endTime, 10))

	resp, err := client.Post(values)

	if err != nil {
		return trades, err
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return trades, fmt.Errorf("request failed: %s\n\n", resp.Status, bytes.NewBuffer(body).String())
	}

	trades = make(map[string][]*PlxPrivateTrade)

	if marketName == "all" {
		err = json.Unmarshal(body, &trades)
	} else {
		marketTrades := make([]*PlxPrivateTrade, 0, 100)
		err = json.Unmarshal(body, &marketTrades)
		trades[marketName] = marketTrades
	}

	return trades, err
}

type TradingApiClient struct {
	*http.Client
}

func NewTradingApiClient() *TradingApiClient {
	client := &http.Client{}
	return &TradingApiClient{client}
}

func (client *TradingApiClient) Post(formData *url.Values) (*http.Response, error) {
	apiKey, apiSecret, err := client.ReadTradingApiCredentials()

	if err != nil {
		return nil, err
	}

	reqBody, signature := client.SignFormData(apiSecret, formData)

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", POLONIEX_TRADING_API_URL, bytes.NewBufferString(reqBody))

	if err != nil {
		return nil, err
	}

	req.Header.Set("Key", apiKey)
	req.Header.Set("Sign", signature)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	return client.Do(req)
}

func (client *TradingApiClient) SignFormData(apiSecret string, formData *url.Values) (reqBody, signature string) {
	nonce := time.Now().UnixNano()
	formData.Set("nonce", strconv.FormatInt(nonce, 10))
	reqBody = formData.Encode()

	mac := hmac.New(sha512.New, []byte(apiSecret))
	mac.Write([]byte(reqBody))
	sigBytes := mac.Sum(nil)
	sigEncoded := hex.EncodeToString(sigBytes)

	return reqBody, sigEncoded
}

func (client *TradingApiClient) ReadTradingApiCredentials() (key, secret string, err error) {
	creds := make(map[string]string)

	fileContent, err := ioutil.ReadFile(API_CREDENTIALS_FILE)
	if err != nil {
		return key, secret, err
	}

	err = json.Unmarshal(fileContent, &creds)
	if err != nil {
		return key, secret, err
	}

	return creds["key"], creds["secret"], nil
}
