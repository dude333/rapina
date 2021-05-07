package reports

import (
	"database/sql"
	"errors"
	"html/template"
	"log"
	"math"
	"net/http"
	"path/filepath"

	"github.com/dude333/rapina/fetch"
	"github.com/dude333/rapina/parsers"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type Server struct {
	db         *sql.DB
	fetchFII   *fetch.FII
	fetchStock *fetch.StockFetch
	log        *Logger
	dataDir    string
	apiKey     string
}

type ServerOption func(*Server)

func WithDB(db *sql.DB) ServerOption {
	return func(s *Server) {
		s.db = db
	}
}
func WithAPIKey(apiKey string) ServerOption {
	return func(s *Server) {
		s.apiKey = apiKey
	}
}
func WithDataDir(dataDir string) ServerOption {
	return func(s *Server) {
		s.dataDir = dataDir
	}
}

func initServer(opts ...ServerOption) (*Server, error) {
	var srv Server
	for _, opt := range opts {
		opt(&srv)
	}
	if srv.db == nil {
		return nil, errors.New("BD inválido")
	}

	srv.db.SetMaxOpenConns(1)
	log := NewLogger(nil)
	stockParser := parsers.NewStock(srv.db, log)
	fiiParser, err := parsers.NewFII(srv.db, log)
	if err != nil {
		return nil, err
	}
	fetchStock := fetch.NewStock(stockParser, log, srv.apiKey, srv.dataDir)
	fetchFII := fetch.NewFII(fiiParser, log)

	srv.log = log
	srv.fetchFII = fetchFII
	srv.fetchStock = fetchStock

	return &srv, nil
}

// HTMLServer based on
// https://www.alexedwards.net/blog/serving-static-sites-with-go
func HTMLServer(codes []string, n int, opts ...ServerOption) {
	srv, err := initServer(opts...)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveTemplate(w, r, srv, codes, n)
	})

	log.Println("Listening on :3000...")
	err = http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func serveTemplate(w http.ResponseWriter, r *http.Request, srv *Server, codes []string, n int) {
	lp := filepath.Join("templates", "layout.html")
	fp := filepath.Join("templates", filepath.Clean(r.URL.Path))
	if fp == "templates" {
		fp = "templates/index.html"
	}

	log.Println("fp:", fp)

	tmpl, err := template.New("").Funcs(template.FuncMap{
		"ptFmtFloat": ptFmtFloat,
	}).ParseFiles(lp, fp)
	if err != nil {
		log.Println(err)
		return
	}

	data := fiiDividends(srv, codes, n)

	err = tmpl.ExecuteTemplate(w, "layout", data)
	if err != nil {
		log.Println(err)
	}
}

type data struct {
	Code   string
	Values []value
}
type value struct {
	Date     string
	Dividend float64
	Quote    float64
	Yeld     float64
	YeldYear float64
}

func fiiDividends(srv *Server, codes []string, n int) *[]data {
	var dataset []data

	for _, code := range codes {
		values := make([]value, 0, n)

		div, err := srv.fetchFII.Dividends(code, n)
		if err != nil {
			srv.log.Error("%s dividendos: %v", code, err)
			return &dataset
		}

		for _, d := range *div {
			q, err := srv.fetchStock.Quote(code, d.Date)
			if err != nil {
				srv.log.Error("Cotação de %s (%s): %v", code, d.Date, err)
				continue
			}

			v := value{
				Date:     d.Date,
				Dividend: d.Val,
				Quote:    q,
			}
			if q > 0 {
				i := d.Val / q
				v.Yeld = 100 * i
				v.YeldYear = 100 * (math.Pow(1+i, 12) - 1)
			}
			values = append(values, v)
		}

		d := data{
			Code:   code,
			Values: values,
		}
		dataset = append(dataset, d)
	}

	return &dataset
}

func ptFmtFloat(f float64) string {
	p := message.NewPrinter(language.BrazilianPortuguese)
	return p.Sprintf("%.2f", f)
}
