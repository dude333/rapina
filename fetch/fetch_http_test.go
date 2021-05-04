package fetch

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

var ts *httptest.Server

func init() {
	handler := http.NewServeMux()
	handler.HandleFunc("/server/api/v1/json", jsonsMock)

	ts = httptest.NewServer(handler)
}

func jsonsMock(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte(`{"text": "mock"}`))
}

type jsonData struct {
	Text string `json:"text"`
}

func TestHTTPFetch_JSON(t *testing.T) {
	h := NewHTTP()

	var got jsonData

	err := h.JSON(ts.URL+"/server/api/v1/json", &got)

	assert.Equal(t, jsonData{Text: "mock"}, got)
	assert.Nil(t, err)
}
