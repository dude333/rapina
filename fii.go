package rapina

import (
	"fmt"

	"github.com/dude333/rapina/parsers"
)

func FetchFII() {
	err := parsers.FetchFII("http://fnet.bmfbovespa.com.br/fnet/publico")
	if err != nil {
		fmt.Println("FetchFII:", err)
	}
}
