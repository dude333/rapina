package fetch

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const apiServer = "https://www.alphavantage.co/"

// Store service used to store data on DB.
type Store interface {
	StockCsv(stream io.ReadCloser, code string) error
}

type StockServer struct {
	apiKey string
	store  Store
}

//
// NewStockServer returns a new instance of *StockServer
//
func NewStockServer(store Store, apiKey string) (*StockServer, error) {
	if store == nil {
		return nil, fmt.Errorf("invalid store service")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("invalid API key: '%s'", apiKey)
	}
	s := StockServer{
		apiKey: apiKey,
		store:  store,
	}
	return &s, nil
}

//
// FetchStockQuote fetches the daily time series (date, daily open, daily high,
// daily low, daily close, daily volume) of the global equity specified,
// covering 20+ years of historical data.
//
func (s StockServer) FetchStockQuote(code string) error {
	tr := &http.Transport{
		DisableCompression: true,
		IdleConnTimeout:    30 * time.Second,
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	v := url.Values{}
	v.Set("function", "TIME_SERIES_DAILY")
	v.Add("symbol", code+".SA")
	v.Add("apikey", s.apiKey)
	v.Add("outputsize", "compact")
	v.Add("datatype", "csv")

	u := JoinURL(apiServer, "query?"+v.Encode())

	fmt.Println(u)

	resp, err := client.Get(u)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %s", resp.Status, u)
	}

	return s.store.StockCsv(resp.Body, code)
}

func (s StockServer) QuoteFromDB(code, date string) (float64, error) {
	return 9.99, nil
}
