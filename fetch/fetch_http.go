package fetch

import (
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

	return json.NewDecoder(r.Body).Decode(target)
}
