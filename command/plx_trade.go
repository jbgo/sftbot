/**
Inputs:
  - plx.GetChartData() for last 24 hours to calculate percentiles
  - plx.GetOpenOrders() what orders do we have on the books waiting to fill?
  - plx.GetCompleteBalances() for currencies in market we plan to trade
  - plx.GetTradeHistory() for last 60 seconds to determine current price

Outputs:
 - trade decision (buy, sell, hold, etc.)
 - trade actions (buy, sell, cancel, etc.)
 - log full API calls in separate log from trading log
 - save current State to DB
 - timeseries data points - for analysis later (these can go in DB)
*/

package command
