package web

import (
	log "github.com/Sirupsen/logrus"
	"net/http"

	"github.com/caascade/posgreSQL/posgresql/executor"
)

const task = "web"

var registration_uuid string

func Init(uuid string, serveDir string) {
	registration_uuid = uuid
	go func() {
		executor.ObtainToken(task, uuid)
		log.Info("Web server is running!")
		serve(serveDir)
		executor.ReturnToken(task, uuid)
	}()
}

func serve(serveDir string) {
	go func() {
		log.Fatal(http.ListenAndServe("0.0.0.0:8081", http.FileServer(http.Dir(serveDir))))
	}()
}
