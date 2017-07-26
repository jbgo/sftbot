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

type CompleteBalance struct {
	Currency  string
	Available float64 `json:"available"`
	OnOrders  float64 `json:"onOrders"`
	BtcValue  float64 `json:"btcValue"`
}

func (client *Client) GetBalance(currency string) (balance *CompleteBalance, err error) {
	balances, err := client.CompleteBalances()
	if err != nil {
		return nil, err
	}

	for _, b := range balances {
		if b.Currency == currency {
			balance = &b
			break
		}
	}

	if balance == nil {
		err = fmt.Errorf("could not find balance for currency: %s", currency)
	}

	return balance, err
}

func (client *Client) CompleteBalances() ([]CompleteBalance, error) {
	balances := make([]CompleteBalance, 0)

	values := &url.Values{}
	values.Set("command", "returnCompleteBalances")

	resp, err := client.TradingApiRequest(values)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		fmt.Printf("HTTP ERROR %s\n%s\n", resp.Status, string(body))
		return nil, fmt.Errorf("request failed: %s\n\n", resp.Status)
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

func (client *Client) AllOpenOrders() (marketOrders map[string][]OpenOrder, err error) {
	marketOrders = make(map[string][]OpenOrder)

	values := &url.Values{}
	values.Set("command", "returnOpenOrders")
	values.Set("currencyPair", "all")

	resp, err := client.TradingApiRequest(values)

	if err != nil {
		return marketOrders, err
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		fmt.Printf("HTTP ERROR %s\n%s\n", resp.Status, string(body))
		return marketOrders, fmt.Errorf("request failed: %s\n\n", resp.Status)
	}

	err = json.Unmarshal(body, &marketOrders)

	return marketOrders, err
}

func (client *Client) GetOpenOrders(currencyPair string) (openOrders []*OpenOrder, err error) {
	values := &url.Values{}
	values.Set("command", "returnOpenOrders")
	values.Set("currencyPair", currencyPair)

	resp, err := client.TradingApiRequest(values)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		fmt.Printf("HTTP ERROR %s\n%s\n", resp.Status, string(body))
		return nil, fmt.Errorf("request failed: %s\n\n", resp.Status)
	}

	openOrders = make([]*OpenOrder, 0)
	err = json.Unmarshal(body, &openOrders)

	return openOrders, err
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

func (client *Client) MyTradeHistory(marketName string, startTime, endTime int64) (trades map[string][]*PlxPrivateTrade, err error) {

	values := &url.Values{}
	values.Set("command", "returnTradeHistory")
	values.Set("currencyPair", marketName)
	values.Set("start", strconv.FormatInt(startTime, 10))
	values.Set("end", strconv.FormatInt(endTime, 10))

	resp, err := client.TradingApiRequest(values)

	if err != nil {
		return trades, err
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		fmt.Printf("HTTP ERROR %s\n%s\n", resp.Status, string(body))
		return trades, fmt.Errorf("request failed: %s\n\n", resp.Status)
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

type PlxOrder struct {
	OrderNumber     int64 `json:",string"`
	Type            string
	CurrencyPair    string
	Rate            float64 `json:",string"`
	Amount          float64 `json:",string"`
	ResultingTrades []*PlxPrivateTrade
}

func (client *Client) Buy(currencyPair string, rate, amount float64) (plxOrder *PlxOrder, err error) {
	return client.PlaceMarketOrder("buy", currencyPair, rate, amount)
}

func (client *Client) Sell(currencyPair string, rate, amount float64) (plxOrder *PlxOrder, err error) {
	return client.PlaceMarketOrder("sell", currencyPair, rate, amount)
}

func (client *Client) PlaceMarketOrder(apiCommand, currencyPair string, rate, amount float64) (plxOrder *PlxOrder, err error) {
	values := &url.Values{}
	values.Set("command", apiCommand)
	values.Set("currencyPair", currencyPair)
	values.Set("rate", strconv.FormatFloat(rate, 'f', -1, 64))
	values.Set("amount", strconv.FormatFloat(amount, 'f', -1, 64))

	resp, err := client.TradingApiRequest(values)

	if err != nil {
		return nil, err
	}

	plxOrder = &PlxOrder{}

	err = decodeJsonResponse(resp, plxOrder, 200, 201)

	return plxOrder, err
}

func decodeJsonResponse(resp *http.Response, value interface{}, expectedStatusCodes ...int) error {
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	success := false
	for _, statusCode := range expectedStatusCodes {
		success = success || resp.StatusCode == statusCode
	}

	if !success {
		fmt.Printf("HTTP ERROR %s\n%s\n", resp.Status, string(body))
		return fmt.Errorf("request failed: %s\n\n", resp.Status)
	}

	return json.Unmarshal(body, &value)
}

func (client *Client) TradingApiRequest(formData *url.Values) (*http.Response, error) {
	apiKey, apiSecret, err := client.ReadTradingApiCredentials()

	if err != nil {
		return nil, err
	}

	reqBody, signature := client.SignFormData(apiSecret, formData)

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", client.TradingApiUrl(), bytes.NewBufferString(reqBody))

	if err != nil {
		return nil, err
	}

	req.Header.Set("Key", apiKey)
	req.Header.Set("Sign", signature)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	httpClient := http.Client{}

	return httpClient.Do(req)
}

func (client *Client) SignFormData(apiSecret string, formData *url.Values) (reqBody, signature string) {
	nonce := time.Now().UnixNano()
	formData.Set("nonce", strconv.FormatInt(nonce, 10))
	reqBody = formData.Encode()

	mac := hmac.New(sha512.New, []byte(apiSecret))
	mac.Write([]byte(reqBody))
	sigBytes := mac.Sum(nil)
	sigEncoded := hex.EncodeToString(sigBytes)

	return reqBody, sigEncoded
}

func (client *Client) ReadTradingApiCredentials() (key, secret string, err error) {
	creds := make(map[string]string)

	fileContent, err := ioutil.ReadFile(client.CredentialsPath)
	if err != nil {
		return key, secret, err
	}

	err = json.Unmarshal(fileContent, &creds)
	if err != nil {
		return key, secret, err
	}

	return creds["key"], creds["secret"], nil
}
