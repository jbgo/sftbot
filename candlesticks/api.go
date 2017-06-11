package candlesticks

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const PUBLIC_API_URL = "https://poloniex.com/public"

type ChartDataParams struct {
	command      string
	currencyPair string
	start        uint
	end          uint
	period       uint
}

type Candlestick struct {
	Date            uint64
	High            float64
	Low             float64
	Open            float64
	Close           float64
	Volume          float64
	QuoteVolume     float64
	WeightedAverage float64
}

func (p *ChartDataParams) ToQueryString() string {
	return fmt.Sprintf(
		"command=%s&currencyPair=%s&start=%d&end=%d&period=%d",
		p.command,
		p.currencyPair,
		p.start,
		p.end,
		p.period)
}

func Get() (string, error) {
	params := ChartDataParams{"returnChartData", "BTC_XRP", 1497033851, 1497120251, 1800}

	url := fmt.Sprintf("%s?%s", PUBLIC_API_URL, params.ToQueryString())

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		log.Panic("request failed:", resp.Status)
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	log.Printf("%d %s GET %s", len(body), resp.Status, url)

	candlesticks := make([]Candlestick, 24)
	err = json.Unmarshal(body, &candlesticks)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
