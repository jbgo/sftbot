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
