package candlesticks

import (
	"encoding/json"
	"fmt"
	"github.com/jbgo/sftbot/data"
	"io/ioutil"
	"log"
	"net/http"
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

func GetChartData(params *ChartDataParams) (*[]data.Candlestick, error) {
	var sticks []data.Candlestick

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
