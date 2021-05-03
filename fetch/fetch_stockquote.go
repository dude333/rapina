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

// API providers
const (
	APInone = iota
	APIalphavantage
	APIyahoo
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

	err = s.stockQuoteFromAPIServer(code, date, APIyahoo)
	if err != nil && s.apiKey != "" {
		// Fallback to Alpha Vantage if Yahoo fails
		s.log.Debug("Cotação não encontrada no Yahoo, tentando no Alpha Vantage")
		err = s.stockQuoteFromAPIServer(code, date, APIalphavantage)
		if err != nil {
			return 0, err
		}
	}

	return s.store.Quote(code, date)
}

//
// stockQuoteFromAPIServer fetches the daily time series (date, daily open, daily high,
// daily low, daily close, daily volume) of the global equity specified,
// covering 20+ years of historical data.
//
func (s StockFetch) stockQuoteFromAPIServer(code, date string, apiProvider int) error {
	if v := s.cache[code]; v == APIalphavantage && apiProvider == APIalphavantage {
		// return fmt.Errorf("cotação histórica para '%s' já foi feita", code)
		return nil // silent return if this fetch has been run already
	}

	s.log.Printf("[>] Baixando cotações de %s\n", code)

	// Download quote for 'code'
	tr := &http.Transport{
		DisableCompression: true,
		IdleConnTimeout:    30 * time.Second,
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	u := apiURL(apiProvider, s.apiKey, code, date)
	if u == "" {
		return errors.New("URL do API server")
	}
	resp, err := client.Get(u)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s", resp.Status)
	}

	s.cache[code] = apiProvider // mark map to avoid unnecessary downloads

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

func apiURL(provider int, apiKey, code, date string) string {
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
		const layout = "2006-01-02 15:04:05 -0700 MST"
		t1, err1 := time.Parse(layout, date+" 00:00:00 -0300 GMT")
		t2, err2 := time.Parse(layout, date+" 23:59:59 -0300 GMT")
		if err1 != nil || err2 != nil {
			return ""
		}
		v.Set("period1", fmt.Sprint(t1.Unix()))
		v.Add("period2", fmt.Sprint(t2.Unix()))
		v.Add("interval", "1d")
		v.Add("events", "history")
		v.Add("includeAdjustedClose", "true")
		return fmt.Sprintf("https://query1.finance.yahoo.com/v7/finance/download/%s.SA?%s",
			code, v.Encode())
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
