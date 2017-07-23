package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	log "github.com/Sirupsen/logrus"

	"github.com/caascade/posgreSQL/posgresql/app"
	"github.com/caascade/posgreSQL/posgresql/executor"
	"github.com/caascade/posgreSQL/posgresql/resource"

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
	r.HandleFunc("/secret", secretHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/address", addressHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/address/{type}", addressHandler).Methods("PUT", "OPTIONS")
	r.HandleFunc("/scale/{scale}", scaleHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/reset-slaves", resetHandler).Methods("PUT", "OPTIONS")
	r.HandleFunc("/state", stateHandler).Methods("GET", "PUT", "OPTIONS")
	r.HandleFunc("/log/master/post", logHandlerPost)
	r.HandleFunc("/log/master/get", logHandlerGet)
	r.HandleFunc("/", handler).Methods("GET", "PUT", "OPTIONS")
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
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, HEAD")
	w.Header().Set("Access-Control-Allow-Headers", "X-PINGOTHER, Content-Type")
	w.Header().Set("Access-Control-Max-Age", "86400")
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

func secretHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, HEAD")
	w.Header().Set("Access-Control-Allow-Headers", "X-PINGOTHER, Content-Type")
	w.Header().Set("Access-Control-Max-Age", "86400")
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	passwd := resource.Password{}
	err = json.Unmarshal(body, &passwd)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}
	status, resp := app.SetPassword(passwd)
	w.WriteHeader(status)
	w.Write([]byte(resp))
}

func scaleHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, HEAD")
	w.Header().Set("Access-Control-Allow-Headers", "X-PINGOTHER, Content-Type")
	w.Header().Set("Access-Control-Max-Age", "86400")
	vars := mux.Vars(r)
	scale := vars["scale"]
	scaleNum, err := strconv.Atoi(scale)
	if err != nil || scaleNum < 1 || scaleNum > 8 {
		w.WriteHeader(500)
		w.Write([]byte("Please specify a valid number in range [1, 8]"))
	}
	status, message := app.ScaleApp(scaleNum)
	w.WriteHeader(status)
	w.Write([]byte(message))
}

func addressHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, HEAD")
	w.Header().Set("Access-Control-Allow-Headers", "X-PINGOTHER, Content-Type")
	w.Header().Set("Access-Control-Max-Age", "86400")
	if r.Method == "GET" {
		status, addresses := app.GetAddresses()
		w.WriteHeader(status)
		w.Write([]byte(addresses))
	} else if r.Method == "PUT" {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		defer r.Body.Close()
		addrs := resource.Address{}
		err = json.Unmarshal(data, &addrs)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		vars := mux.Vars(r)
		status, resp := app.UpdateAddresses(addrs, vars["type"])
		w.WriteHeader(status)
		w.Write([]byte(resp))
	}
}

func resetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, HEAD")
	w.Header().Set("Access-Control-Allow-Headers", "X-PINGOTHER, Content-Type")
	w.Header().Set("Access-Control-Max-Age", "86400")
	respCode, msg := app.ResetSlaves()
	w.WriteHeader(respCode)
	w.Write([]byte(msg))
}

func stateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, HEAD")
	w.Header().Set("Access-Control-Allow-Headers", "X-PINGOTHER, Content-Type")
	w.Header().Set("Access-Control-Max-Age", "86400")
	if r.Method == "GET" {
		appl := app.GetApp()
		w.WriteHeader(200)
		w.Write([]byte(appl.Status.State))
	}
	if r.Method == "PUT" {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		defer r.Body.Close()
		status, x := app.UpdateState(string(data))
		w.WriteHeader(status)
		w.Write([]byte(x))
	}
}
