package plx

import (
	"encoding/json"
	"fmt"
	"github.com/jbgo/sftbot/data"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

const PUBLIC_API_URL = "https://poloniex.com/public"

type ChartDataParams struct {
	CurrencyPair string
	Start        int64
	End          int64
	Period       int64
}

func (p *ChartDataParams) ToQueryString() string {
	return fmt.Sprintf(
		"command=%s&currencyPair=%s&start=%d&end=%d&period=%d",
		"returnChartData",
		p.CurrencyPair,
		p.Start,
		p.End,
		p.Period)
}

func GetChartData(params *ChartDataParams) (*[]data.ChartData, error) {
	var sticks []data.ChartData

	url := fmt.Sprintf("%s?%s", PUBLIC_API_URL, params.ToQueryString())

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		log.Panic("request failed:", resp.Status)
		return nil, fmt.Errorf("request failed: %s", resp.Status)
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	err = json.Unmarshal(body, &sticks)
	if err != nil {
		return nil, err
	}

	return &sticks, nil
}

type TradeHistoryParams struct {
	Market    string
	StartTime int64
	EndTime   int64
}

type Trade struct {
	Date   time.Time
	Type   string
	Rate   float64
	Amount float64
	Total  float64
}

type PlxTrade struct {
	Date   string
	Type   string
	Rate   string
	Amount string
	Total  string
}

func (p *TradeHistoryParams) ToQueryString() string {
	return fmt.Sprintf("command=%s&currencyPair=%s&start=%d&end=%d",
		"returnTradeHistory", p.Market, p.StartTime, p.EndTime)
}

func GetTradeHistory(params *TradeHistoryParams) (trades []Trade, err error) {
	url := fmt.Sprintf("%s?%s", PUBLIC_API_URL, params.ToQueryString())

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("request failed: %s", resp.Status)
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	respData := make([]PlxTrade, 0, 1024)
	err = json.Unmarshal(body, &respData)
	if err != nil {
		return nil, err
	}

	for _, d := range respData {
		trade := Trade{}

		trade.Date, _ = time.Parse("2006-01-02 15:04:05", d.Date)
		trade.Type = d.Type
		trade.Rate, _ = strconv.ParseFloat(d.Rate, 64)
		trade.Amount, _ = strconv.ParseFloat(d.Amount, 64)
		trade.Total, _ = strconv.ParseFloat(d.Total, 64)

		trades = append(trades, trade)
	}

	return trades, nil
}

type TickerEntry struct {
	Market        string
	Last          float64 `json:",string"`
	LowestAsk     float64 `json:",string"`
	HighestBid    float64 `json:",string"`
	PercentChange float64 `json:",string"`
	BaseVolume    float64 `json:",string"`
	QuoteVolume   float64 `json:",string"`
}

func GetTicker() (ticker []TickerEntry, err error) {
	url := fmt.Sprintf("%s?command=returnTicker", PUBLIC_API_URL)

	resp, err := http.Get(url)
	if err != nil {
		return ticker, err
	}

	if resp.StatusCode != 200 {
		return ticker, fmt.Errorf("request failed: %s", resp.Status)
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	respData := make(map[string]TickerEntry)
	err = json.Unmarshal(body, &respData)
	if err != nil {
		return ticker, err
	}

	for k, v := range respData {
		v.Market = k
		ticker = append(ticker, v)
	}

	return ticker, err
}
