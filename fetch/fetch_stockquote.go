package fetch

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/dude333/rapina"
	"github.com/pkg/errors"
)

type StockFetch struct {
	apiKey string
	store  rapina.StockStore
	cache  map[string]int
	log    rapina.Logger
}

//
// NewStockFetch returns a new instance of *StockServer
//
func NewStockFetch(store rapina.StockStore, log rapina.Logger, apiKey string) *StockFetch {
	return &StockFetch{
		apiKey: apiKey,
		store:  store,
		cache:  make(map[string]int),
		log:    log,
	}
}

// Quote returns the quote for 'code' on 'date'.
// Date format: YYYY-MM-DD.
func (s StockFetch) Quote(code, date string) (float64, error) {
	if !rapina.IsDate(date) {
		return 0, fmt.Errorf("data inválida: %s", date)
	}

	val, err := s.store.Quote(code, date)
	if err == nil {
		return val, nil // returning data found on db
	}

	err = s.stockQuoteFromAPIServer(code)
	if err != nil {
		return 0, err
	}

	return s.store.Quote(code, date)
}

//
// stockQuoteFromAPIServer fetches the daily time series (date, daily open, daily high,
// daily low, daily close, daily volume) of the global equity specified,
// covering 20+ years of historical data.
//
func (s StockFetch) stockQuoteFromAPIServer(code string) error {
	if _, ok := s.cache[code]; ok {
		return fmt.Errorf("cotação histórica para '%s' já foi feita", code)
	}

	s.log.Printf("[>] Baixando cotações de %s\n", code)

	// Download quote for 'code'
	tr := &http.Transport{
		DisableCompression: true,
		IdleConnTimeout:    30 * time.Second,
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	u := apiURL(APIalphavantage, code, s.apiKey)
	s.log.Debug("%s", u)
	resp, err := client.Get(u)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %s", resp.Status, u)
	}

	s.cache[code] += 1 // mark map to avoid unnecessary downloads

	// JSON means error response
	if resp.Header.Get("Content-Type") == "application/json" {
		jsonMap := make(map[string]interface{})
		err := json.NewDecoder(resp.Body).Decode(&jsonMap)
		if err != nil {
			return err
		}
		return errors.New(map2str(jsonMap))
	}

	s.log.Run("Armazendo cotações no banco de dados...")
	_, err = s.store.Save(resp.Body, code)
	if err != nil {
		s.log.Nok()
		return errors.Wrapf(err, "armazenando cotações de %s", code)
	}
	s.log.Ok()

	return err
}

const (
	APIalphavantage int = iota + 1
	APIyahoo
)

func apiURL(provider int, code, apiKey string) string {
	v := url.Values{}
	switch provider {
	case APIalphavantage:
		v.Set("function", "TIME_SERIES_DAILY")
		v.Add("symbol", code+".SA")
		v.Add("apikey", apiKey)
		v.Add("outputsize", "full")
		v.Add("datatype", "csv")
		return "https://www.alphavantage.co/query?" + v.Encode()

	case APIyahoo:
		return ""
	}

	return ""
}

func map2str(data map[string]interface{}) string {
	var buf string
	for k, v := range data {
		buf += fmt.Sprintln(k+":", v)
	}
	return buf
}
