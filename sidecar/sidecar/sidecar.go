package sidecar

import (
	"fmt"
	"net/http"

	"github.com/caascade/posgreSQL/sidecar/cmdline"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

func InitSidecar(input *cmdline.CmdlineArgs) {
	r := mux.NewRouter()
	r.HandleFunc("/", handler).Methods("GET")
	srv := &http.Server{
		Handler: r,
		Addr:    fmt.Sprintf("%s:%d", "0.0.0.0", 9898),
	}
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatalf("Error running server %s", err.Error())
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("ok"))
}
