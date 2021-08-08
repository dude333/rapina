package server

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
	"github.com/dude333/rapina/progress"
	"github.com/dude333/rapina/reports"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type Server struct {
	db         *sql.DB
	fetchFII   *fetch.FII
	fetchStock *fetch.Stock
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
	log := reports.NewLogger(os.Stderr)
	fetchStock, err := fetch.NewStock(srv.db, log, srv.apiKey, srv.dataDir)
	if err != nil {
		return nil, err
	}
	fetchFII, err := fetch.NewFII(srv.db, log)
	if err != nil {
		return nil, err
	}

	srv.fetchFII = fetchFII
	srv.fetchStock = fetchStock

	return &srv, nil
}

// HTML is a very basic html server to handle the reports.
func HTML(opts ...ServerOption) {
	srv, err := initServer(opts...)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", renderTemplate(srv))

	log.Println("Listening on :3000...")
	err = http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatal(err)
	}
}

// renderTemplate renders the file related to the URL path inside the layout
// templates. Template files are locates in _contentFS.
func renderTemplate(srv *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fp := filepath.Clean(r.URL.Path)
		if strings.HasPrefix(fp, `/`) || strings.HasPrefix(fp, `\`) {
			fp = fp[1:] // remove starting "/" (or "\" on Windows)
		}
		if fp == "" {
			fp = "index.html"
		}

		log.Println("rendering", fp)

		// TODO: load all templates outside this funcion
		tmpl, err := template.New("").Funcs(template.FuncMap{
			"ptFmtFloat": ptFmtFloat,
		}).ParseFS(_contentFS, "layout.html", fp)

		if err != nil {
			log.Println(err)
			return
		}

		// Set the payload according to the URL path
		var payload interface{}
		if strings.Contains(fp, "fii.html") && r.Method == http.MethodPost {
			payload = payloadFIIDividends(srv, r)
		}

		err = tmpl.ExecuteTemplate(w, "layout", payload)
		if err != nil {
			log.Println(err)
		}
	}
}

func payloadFIIDividends(srv *Server, r *http.Request) interface{} {
	var payload struct {
		Codes  string
		Months int
		Data   interface{}
	}

	codes := parseCodes(r.FormValue("codes"))
	months, err := strconv.Atoi(r.FormValue("months"))
	if err != nil {
		months = 1
	}
	payload.Codes = strings.Join(codes, " ")
	payload.Months = months
	payload.Data = fiiDividends(srv, codes, months)

	return &payload
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
			progress.ErrorMsg("%s: %v", code, err)
			continue
		}

		for _, d := range *div {
			q, err := srv.fetchStock.Quote(code, d.Date)
			if err != nil {
				progress.ErrorMsg("Cotação de %s (%s): %v", code, d.Date, err)
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
