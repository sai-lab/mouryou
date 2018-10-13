package monitor

import (
	"net"
	"net/http"

	"fmt"

	"github.com/sai-lab/mouryou/lib/check"
	"github.com/sai-lab/mouryou/lib/models"
)

func MeasureServer(config *models.Config) {
	listener, err := net.Listen("tcp", ":9999")
	check.Error(err)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"message": "OK"}`)
		return
	})
	go http.Serve(listener, nil)
}
