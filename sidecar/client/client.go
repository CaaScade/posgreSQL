package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/caascade/posgreSQL/posgresql/resource"

	log "github.com/Sirupsen/logrus"
)

func UpdateMasterAddress(controllerIP string, controllerPort int) {
	selfIP := os.Getenv("SELF_IP")
	if selfIP == "" {
		log.Fatalf("SELF_IP is not set, cannot proceed")
	}

	resp, err := http.Get(fmt.Sprintf("http://%s/", controllerIP))
	if err != nil {
		log.Fatalf("error getting app obj: %s", err.Error())
	}

	var appl resource.Application

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error reading resp body %s", err.Error())
	}
	defer resp.Body.Close()

	err = json.Unmarshal(body, &appl)
	if err != nil {
		log.Fatalf("error unmarshalling app obj: %s", err.Error())
	}

	appl.Status.Addresses.MasterIP = selfIP

	data, err := json.Marshal(appl)
	if err != nil {
		log.Fatalf("Error marshalling addresses %s", err.Error())
	}

	clientx := http.Client{}
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("http://%s/", controllerIP), bytes.NewBuffer(data))
	if err != nil {
		log.Fatalf("Error updating master ip %s", err.Error())
	}
	resp, err = clientx.Do(req)
	if err != nil {
		log.Fatalf("Error updating master ip %s", err.Error())
	}
	if resp.StatusCode != 200 {
		log.Fatalf("got non-200 status code on update. Exiting")
	}
}

func UpdateSlaveAddress(controllerIP string, controllerPort int) {
	selfIP := os.Getenv("SELF_IP")
	if selfIP == "" {
		log.Fatalf("SELF_IP is not set, cannot proceed")
	}
	resp, err := http.Get(fmt.Sprintf("http://%s/", controllerIP))
	if err != nil {
		log.Fatalf("error getting app obj: %s", err.Error())
	}

	var appl resource.Application

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error reading resp body %s", err.Error())
	}
	defer resp.Body.Close()

	err = json.Unmarshal(body, &appl)
	if err != nil {
		log.Fatalf("error unmarshalling app obj: %s", err.Error())
	}

	appl.Status.Addresses.SlaveIP = selfIP

	data, err := json.Marshal(appl)
	if err != nil {
		log.Fatalf("Error marshalling addresses %s", err.Error())
	}

	clientx := http.Client{}
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("http://%s/", controllerIP), bytes.NewBuffer(data))
	if err != nil {
		log.Fatalf("Error updating slave ip %s", err.Error())
	}
	resp, err = clientx.Do(req)
	if err != nil {
		log.Fatalf("Error updating slave ip %s", err.Error())
	}

	if resp.StatusCode != 200 {
		log.Fatalf("got non-200 status code on update. Exiting")
	}
}

func GetSlaveAddress(controllerIP string, controllerPort int) (string, int) {
	log.Errorf("Quering for slave address")
	//resp, err := http.Get(fmt.Sprintf("http://%s:%d/address", controllerIP, controllerPort))
	resp, err := http.Get(fmt.Sprintf("http://%s/address", controllerIP))
	if err != nil {
		log.Fatalf("error getting slave address %s", err.Error())
	}
	var addresses resource.Addresses
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Fatalf("Error reading resp body %s", err.Error())
	}
	err = json.Unmarshal(body, &addresses)
	if err != nil {
		log.Fatalf("resp body cannot be unmarshalled %s", err.Error())
	}
	if addresses.SlaveIP == "" {
		log.Fatalf("slave IP is empty")
	}
	return addresses.SlaveIP, addresses.SlavePort
}

func GetMasterAddress(controllerIP string, controllerPort int) (string, int) {
	log.Errorf("Quering for master address")
	//resp, err := http.Get(fmt.Sprintf("http://%s:%d/address", controllerIP, controllerPort))
	resp, err := http.Get(fmt.Sprintf("http://%s/address", controllerIP))
	if err != nil {
		log.Fatalf("error getting slave address %s", err.Error())
	}
	var addresses resource.Addresses
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Fatalf("Error reading resp body %s", err.Error())
	}
	err = json.Unmarshal(body, &addresses)
	if err != nil {
		log.Fatalf("resp body cannot be unmarshalled %s", err.Error())
	}
	if addresses.MasterIP == "" {
		log.Fatalf("master IP is empty")
	}
	return addresses.MasterIP, addresses.MasterPort
}
