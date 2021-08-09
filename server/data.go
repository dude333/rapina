package server

import (
	"math"
	"net/url"
	"strings"

	"github.com/dude333/rapina/progress"
)

// fiiDividendsPayload returns the data to be used in the FII template.
func fiiDividendsPayload(srv *Server, codes []string, months int) interface{} {
	var payload struct {
		Codes  string
		Months int
		Data   interface{}
	}

	payload.Codes = strings.Join(codes, " ")
	payload.Months = months
	payload.Data = fiiDividends(srv, codes, months)

	return &payload
}

func fiiDividends(srv *Server, codes []string, n int) interface{} {
	type value struct {
		Date     string
		Dividend float64
		Quote    float64
		Yeld     float64
		YeldYear float64
	}

	type data struct {
		Code    string
		Name    string
		Website string
		Values  []value
	}

	var dataset []data

	// Fill 'data' for every stock code
	for _, code := range codes {
		code = strings.ToUpper(code)
		values := make([]value, 0, n)

		// Dividends from last "n" months
		div, err := srv.fetchFII.Dividends(code, n)
		if err != nil {
			progress.ErrorMsg("%s: %v", code, err)
			continue
		}

		// Stock quotes from the days when the dividends were received
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
	} // next code

	return &dataset
}
