package fetch

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/dude333/rapina"
	"github.com/pkg/errors"
)

const apiServer = "https://www.alphavantage.co/"

type StockFetch struct {
	apiKey string
	store  rapina.StockStore
}

//
// NewStockFetch returns a new instance of *StockServer
//
func NewStockFetch(store rapina.StockStore, apiKey string) (*StockFetch, error) {
	if store == nil {
		return nil, fmt.Errorf("invalid store service")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("invalid API key: '%s'", apiKey)
	}
	s := StockFetch{
		apiKey: apiKey,
		store:  store,
	}
	return &s, nil
}

// Quote returns the quote for 'code' on 'date'.
// Date format: YYYY-MM-DD.
func (s StockFetch) Quote(code, date string) (rapina.Quotation, error) {
	if !rapina.IsDate(date) {
		return rapina.Quotation{}, rapina.ErrInvalidDate
	}

	return s.store.Quote(code, date)
}

//
// FetchStockQuote fetches the daily time series (date, daily open, daily high,
// daily low, daily close, daily volume) of the global equity specified,
// covering 20+ years of historical data.
//
func (s StockFetch) FetchStockQuote(code string) error {
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

	u := rapina.JoinURL(apiServer, "query?"+v.Encode())

	fmt.Print("[ ] Baixando cotações...")

	resp, err := client.Get(u)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("\r[x")
		return fmt.Errorf("%s: %s", resp.Status, u)
	}
	fmt.Println("\r[✓")

	fmt.Print("[ ] Armazendo cotações no banco de dados...")
	err = s.store.CsvToDB(resp.Body, code)
	if err != nil {
		fmt.Println("\r[x")
		return errors.Wrap(err, "armazenando cotações")
	}
	fmt.Println("\r[✓")

	return err
}

func (s StockFetch) QuoteFromDB(code, date string) (float64, error) {
	return 9.99, nil
}
