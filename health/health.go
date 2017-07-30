package health

import (
	"log"
	"net/http"
)


func Run() {
	// Handler for returning health check on /healthz endpoint
	// Started in main as a go routine
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte("ok"))

	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
