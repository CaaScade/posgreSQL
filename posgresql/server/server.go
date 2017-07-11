package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"

	"github.com/caascade/posgreSQL/posgresql/app"
	"github.com/caascade/posgreSQL/posgresql/executor"

	"github.com/gorilla/mux"
)

const task = "server"

var registration_uuid string

func Init(uuid string, listenAddr string, listenPort int) {
	registration_uuid = uuid
	go func() {
		executor.ObtainToken(task, uuid)
		log.Info("Controller is running!")
		serve(listenAddr, listenPort)
		executor.ReturnToken(task, uuid)
	}()
}

func serve(addr string, port int) {
	r := mux.NewRouter()
	r.HandleFunc("/", handler).Methods("GET", "PUT")
	srv := &http.Server{
		Handler: r,
		Addr:    fmt.Sprintf("%s:%d", addr, port),
	}
	err := srv.ListenAndServe()
	if err != nil {
		executor.SetErrorState(registration_uuid, err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		appInstance := app.GetApp()
		data, _ := json.Marshal(appInstance)
		w.Write(data)
	}
	if r.Method == "PUT" {
		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		status, resp := app.UpdateApp(body)
		w.WriteHeader(status)
		w.Write([]byte(resp))
	}
}
