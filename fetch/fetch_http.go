package fetch

import (
	"encoding/json"
	"net/http"
	"time"
)

type HTTPFetch struct {
	client *http.Client
}

func NewHTTPFetch() *HTTPFetch {
	c := &http.Client{Timeout: 10 * time.Second}
	return &HTTPFetch{client: c}
}

func (h HTTPFetch) JSON(url string, target interface{}) error {
	r, err := h.client.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}
