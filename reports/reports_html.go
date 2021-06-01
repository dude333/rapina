package reports

import (
	"database/sql"
	"errors"
	"html/template"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dude333/rapina/fetch"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type Server struct {
	db         *sql.DB
	fetchFII   *fetch.FII
	fetchStock *fetch.Stock
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
	log := NewLogger(os.Stderr)
	fetchStock, err := fetch.NewStock(srv.db, log, srv.apiKey, srv.dataDir)
	if err != nil {
		return nil, err
	}
	fetchFII, err := fetch.NewFII(srv.db, log)
	if err != nil {
		return nil, err
	}

	srv.log = log
	srv.fetchFII = fetchFII
	srv.fetchStock = fetchStock

	return &srv, nil
}

// HTMLServer is a very basic html server to show the reports.
func HTMLServer(opts ...ServerOption) {
	srv, err := initServer(opts...)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveTemplate(w, r, srv)
	})

	log.Println("Listening on :3000...")
	err = http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func serveTemplate(w http.ResponseWriter, r *http.Request, srv *Server) {
	lp := filepath.Join("templates", "layout.html")
	fp := filepath.Join("templates", filepath.Clean(r.URL.Path))
	if fp == "templates" {
		fp = "templates/index.html"
	}

	log.Println("fp:", fp)

	tmpl, err := template.New("").Funcs(template.FuncMap{
		"ptFmtFloat": ptFmtFloat,
	}).ParseFS(_fs, lp, fp)
	if err != nil {
		log.Println(err)
		return
	}

	var payload struct {
		Codes  string
		Months int
		Data   interface{}
	}
	if strings.Contains(fp, "fii.html") && r.Method == http.MethodPost {
		codes := parseCodes(r.FormValue("codes"))
		months, err := strconv.Atoi(r.FormValue("months"))
		if err != nil {
			months = 1
		}
		payload.Codes = strings.Join(codes, " ")
		payload.Months = months
		payload.Data = fiiDividends(srv, codes, months)
	}

	err = tmpl.ExecuteTemplate(w, "layout", payload)
	if err != nil {
		log.Println(err)
	}
}

type data struct {
	Code    string
	Name    string
	Website string
	Values  []value
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
		code = strings.ToUpper(code)
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

		// FII details, if found
		details, err := srv.fetchFII.Details(code)
		var name, a string
		if err == nil {
			name = details.DetailFund.CompanyName
			u, err := url.Parse(details.DetailFund.WebSite)
			if err == nil && u.Scheme == "" {
				u.Scheme = "https"
				a = u.String()
			}
		}

		d := data{
			Code:    code,
			Name:    name,
			Website: a,
			Values:  values,
		}

		dataset = append(dataset, d)
	}

	return &dataset
}

func ptFmtFloat(f float64) string {
	p := message.NewPrinter(language.BrazilianPortuguese)
	return p.Sprintf("%.2f", f)
}

func split(r rune) bool {
	return r == ' ' || r == ',' || r == ';' || r == '\n'
}

func parseCodes(text string) []string {
	var codes []string
	for _, field := range strings.FieldsFunc(text, split) {
		field = strings.TrimSpace(field)
		if len(field) == len("ABCD11") {
			codes = append(codes, field)
		}
	}

	return codes
}
