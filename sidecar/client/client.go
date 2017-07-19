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
	UpdateAddress("master", controllerIP, controllerPort)
}

func UpdateSlaveAddress(controllerIP string, controllerPort int) {
	UpdateAddress("slave", controllerIP, controllerPort)
}

func ResetSlaves(controllerIP string, controllerPort int) {

	clientx := http.Client{}
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("http://%s/reset-slaves", controllerIP), nil)
	if err != nil {
		log.Fatalf("Error updating master ip %s", err.Error())
	}
	resp, err := clientx.Do(req)
	if err != nil {
		log.Fatalf("Error updating master ip %s", err.Error())
	}
	if resp.StatusCode != 200 {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Error reading resp body %s", err.Error())
		}
		log.Fatalf("got non-200 status code on update. Exiting Status_Code=%d Msg=%s", resp.StatusCode, string(data))
	}

}

func UpdateAddress(addrType, controllerIP string, controllerPort int) {
	selfIP := os.Getenv("SELF_IP")
	if selfIP == "" {
		log.Fatalf("SELF_IP is not set, cannot proceed")
	}

	addr := resource.Address{
		IP: selfIP,
	}

	data, err := json.Marshal(addr)
	if err != nil {
		log.Fatalf("Error marshalling addresses %s", err.Error())
	}

	clientx := http.Client{}
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("http://%s/address/%s", controllerIP, addrType), bytes.NewBuffer(data))
	if err != nil {
		log.Fatalf("Error updating %s address %s", addrType, err.Error())
	}
	resp, err := clientx.Do(req)
	if err != nil {
		log.Fatalf("Error updating %s address %s", addrType, err.Error())
	}
	if resp.StatusCode != 200 {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Error reading resp body %s", err.Error())
		}
		log.Fatalf("got non-200 status code on update. Exiting Status_Code=%d Msg=%s", resp.StatusCode, string(data))
	}
}

func GetMasterAddress(controllerIP string, controllerPort int) (string, int) {
	log.Errorf("Querying for master address")
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
	if addresses.Master.IP == "" {
		log.Fatalf("master IP is empty")
	}
	return addresses.Master.IP, addresses.Master.Port
}
