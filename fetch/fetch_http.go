package fetch

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"time"
)

// HTTPFetch implements a generic HTTP fetcher.
type HTTPFetch struct {
	client *http.Client
}

// NewHTTP creates a new HTTPFetch instance.
func NewHTTP() *HTTPFetch {
	c := &http.Client{Timeout: 10 * time.Second}
	return &HTTPFetch{client: c}
}

// JSON handles json responses.
func (h HTTPFetch) JSON(url string, target interface{}) error {
	r, err := h.client.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	// for _, c := range r.Cookies() {
	// 	fmt.Printf("COOKIE: %+v\n", c)
	// }

	return json.NewDecoder(r.Body).Decode(target)
}

func getJSON(url string, target interface{}) error {
	c := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			DisableCompression: true,
			IdleConnTimeout:    30 * time.Second,
			TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
		},
	}

	r, err := c.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}
