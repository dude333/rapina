package server

import (
	"database/sql"
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dude333/rapina/fetch"
	"github.com/dude333/rapina/reports"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type Server struct {
	db         *sql.DB
	fetchFII   *fetch.FII
	fetchStock *fetch.Stock
	report     *reports.Report
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
	report, err := reports.New(map[string]interface{}{"db": srv.db})
	if err != nil {
		return nil, err
	}

	srv.fetchFII = fetchFII
	srv.fetchStock = fetchStock
	srv.report = report

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
			codes := parseCodes(r.FormValue("codes"))
			months := parseNumeric(r.FormValue("months"), 1)
			payload = fiiDividendsPayload(srv, codes, months)
		}

		err = tmpl.ExecuteTemplate(w, "layout", payload)
		if err != nil {
			log.Println(err)
		}
	}
}

func ptFmtFloat(f float64) string {
	p := message.NewPrinter(language.BrazilianPortuguese)
	return p.Sprintf("%.2f", f)
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

func split(r rune) bool {
	return r == ' ' || r == ',' || r == ';' || r == '\n'
}

// parseNumeric converts "numeric" to integer, or returns "alt" in case of error.
func parseNumeric(numeric string, alt int) int {
	n, err := strconv.Atoi(numeric)
	if err != nil {
		n = alt
	}
	return n
}
