package command

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/jbgo/sftbot/db"
	"github.com/jbgo/sftbot/plx"
	"github.com/jbgo/sftbot/trading"
	"io/ioutil"
	"log"
	"sort"
	"strings"
)

const PROSPECT_DB = "sftbot-prospect.db"

type ProspectCommand struct {
	Flags *flag.FlagSet

	Config string

	DBStore db.Store
}

func (c *ProspectCommand) Synopsis() string {
	return "search for markets that look tradeable"
}

func (c *ProspectCommand) Help() string {
	return formatHelpText(`
Usage: sftbot plx ticker [options]

  ` + c.Synopsis() + `

` + helpOptions(c))
}

func (c *ProspectCommand) InitFlags() *flag.FlagSet {
	c.Flags = flag.NewFlagSet("plx ticker", flag.ContinueOnError)

	c.Flags.StringVar(&c.Config, "config", "", "Trader config file (JSON)")

	return c.Flags
}

type ByVolumeDesc []plx.TickerEntry

func (a ByVolumeDesc) Len() int           { return len(a) }
func (a ByVolumeDesc) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByVolumeDesc) Less(i, j int) bool { return a[i].BaseVolume > a[j].BaseVolume }

func (c *ProspectCommand) Run(args []string) int {
	c.InitFlags()
	c.Flags.Parse(args)
	c.InitDB()

	client := plx.NewLiveClient()
	ticker, err := client.GetTicker()

	if err != nil {
		log.Panic(err)
	}

	sort.Sort(ByVolumeDesc(ticker))

	for _, t := range ticker {
		if t.BaseVolume < 100 {
			continue
		}

		trader, err := c.InitTrader(t.Market)

		if err != nil {
			log.Println("!!! ERROR", err)
			continue
		}

		marketData, err := trader.Prospect()

		if err != nil {
			log.Println("!!! ERROR", err)
			continue
		}

		fmt.Printf(strings.Join([]string{
			fmt.Sprintf("%s", t.Market),
			fmt.Sprintf("price=%0.9f", marketData.CurrentPrice),
			fmt.Sprintf("bid=%0.9f", marketData.Percentiles[trader.BuyThreshold]),
			fmt.Sprintf("volatility=%0.4f", marketData.VolatilityIndex),
			fmt.Sprintf("volume=%0.4f", t.BaseVolume),
			"\n",
		}, " "))
	}

	return 0
}

func (c *ProspectCommand) InitDB() {
	dbPath := LIVE_DB

	dbStore, err := db.NewBoltStore("prospect", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	c.DBStore = dbStore
}

func (c *ProspectCommand) InitTrader(marketName string) (*trading.Trader, error) {
	plxClient := plx.NewLiveClient()
	plxExchange := trading.NewPlxExchange(plxClient)

	traderConfig, err := c.LoadTraderConfig()
	if err != nil {
		return nil, err
	}

	trader, err := trading.NewTrader(marketName, plxExchange, c.DBStore, traderConfig)
	if err != nil {
		return nil, err
	}

	return trader, nil
}

func (c *ProspectCommand) LoadTraderConfig() (*trading.TraderConfig, error) {
	if len(c.Config) == 0 {
		return nil, fmt.Errorf("-config is required")
	}

	data, err := ioutil.ReadFile(c.Config)
	if err != nil {
		return nil, err
	}

	traderConfig := &trading.TraderConfig{}
	err = json.Unmarshal(data, traderConfig)

	return traderConfig, err
}
