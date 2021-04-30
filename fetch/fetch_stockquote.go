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
	cache  map[string]bool
	log    rapina.Logger
}

//
// NewStockFetch returns a new instance of *StockServer
//
func NewStockFetch(store rapina.StockStore, log rapina.Logger, apiKey string) *StockFetch {
	return &StockFetch{
		apiKey: apiKey,
		store:  store,
		cache:  make(map[string]bool),
		log:    log,
	}
}

// Quote returns the quote for 'code' on 'date'.
// Date format: YYYY-MM-DD.
func (s StockFetch) Quote(code, date string) (float64, error) {
	if !rapina.IsDate(date) {
		return 0, rapina.ErrInvalidDate
	}

	val, err := s.stockQuoteFromDB(code, date)
	if err == nil {
		return val, nil
	}

	s.log.Debug("FROM SERVER")
	err = s.stockQuoteFromServer(code)
	if err != nil {
		return 0, err
	}

	return s.stockQuoteFromDB(code, date)
}

//
// stockQuoteFromServer fetches the daily time series (date, daily open, daily high,
// daily low, daily close, daily volume) of the global equity specified,
// covering 20+ years of historical data.
//
func (s StockFetch) stockQuoteFromServer(code string) error {
	if _, ok := s.cache[code]; ok {
		return fmt.Errorf("cotação histórica para '%s' já foi feita", code)
	}
	s.cache[code] = true

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
	v.Add("outputsize", "full")
	v.Add("datatype", "csv")

	u := rapina.JoinURL(apiServer, "query?"+v.Encode())

	s.log.Run("[ ] Baixando cotações de %v", u)

	resp, err := client.Get(u)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.log.Nok()
		return fmt.Errorf("%s: %s", resp.Status, u)
	}
	s.log.Ok()

	s.log.Run("[ ] Armazendo cotações no banco de dados...")
	err = s.store.CsvToDB(resp.Body, code)
	if err != nil {
		s.log.Nok()
		return errors.Wrap(err, "armazenando cotações")
	}
	s.log.Ok()

	return err
}

func (s StockFetch) stockQuoteFromDB(code, date string) (float64, error) {
	return s.store.Quote(code, date)
}
