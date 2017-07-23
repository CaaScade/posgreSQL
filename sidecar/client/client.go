package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/caascade/posgreSQL/posgresql/resource"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
)

func StreamLogs(controllerIP string, controllerPort int, logChan <-chan string) {
	u := url.URL{Scheme: "ws", Host: fmt.Sprintf("%s:%d", controllerIP, controllerPort), Path: "/log/master/post"}

	c, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		data, erx := ioutil.ReadAll(resp.Body)
		if erx != nil {
			log.Errorf("Error reading resp %d %s", resp.StatusCode, erx.Error())
			return
		}
		defer resp.Body.Close()
		log.Errorf("dial url:%s :%s %s", u.String(), err.Error(), string(data))
		return
	}
	defer c.Close()
	for l := range logChan {
		err := c.WriteMessage(websocket.TextMessage, []byte(l))
		if err != nil {
			log.Errorf("Error writing to log reader: %s", err.Error())
			break
		}
	}
}

func UpdateMasterAddressNoPanic(controllerIP string, controllerPort int) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("%s", r)
		}
	}()
	UpdateAddress("master", controllerIP, controllerPort)
}

func UpdateSlaveAddressNoPanic(controllerIP string, controllerPort int) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("%s", r)
		}
	}()
	UpdateAddress("slave", controllerIP, controllerPort)
}

func UpdateMasterAddress(controllerIP string, controllerPort int) {
	defer func() {
		if r := recover(); r != nil {
			log.Fatalf("%s", r)
		}
	}()
	UpdateAddress("master", controllerIP, controllerPort)
}

func UpdateSlaveAddress(controllerIP string, controllerPort int) {
	defer func() {
		if r := recover(); r != nil {
			log.Fatalf("%s", r)
		}
	}()
	UpdateAddress("slave", controllerIP, controllerPort)
}

func ResetSlaves(controllerIP string, controllerPort int) {

	clientx := http.Client{}
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("http://%s:%d/reset-slaves", controllerIP, controllerPort), nil)
	if err != nil {
		panic(fmt.Sprintf("Error updating master ip %s", err.Error()))
	}
	resp, err := clientx.Do(req)
	if err != nil {
		panic(fmt.Sprintf("Error updating master ip %s", err.Error()))
	}
	if resp.StatusCode != 200 {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(fmt.Sprintf("Error reading resp body %s", err.Error()))
		}
		panic(fmt.Sprintf("got non-200 status code on update. Exiting Status_Code=%d Msg=%s", resp.StatusCode, string(data)))
	}

}

func UpdateAddress(addrType, controllerIP string, controllerPort int) {
	selfIP := os.Getenv("SELF_IP")
	if selfIP == "" {
		panic(fmt.Sprintf("SELF_IP is not set, cannot proceed"))
	}

	addr := resource.Address{
		IP: selfIP,
	}

	data, err := json.Marshal(addr)
	if err != nil {
		panic(fmt.Sprintf("Error marshalling addresses %s", err.Error()))
	}

	clientx := http.Client{}
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("http://%s:%d/address/%s", controllerIP, controllerPort, addrType), bytes.NewBuffer(data))
	if err != nil {
		panic(fmt.Sprintf("Error updating %s address %s", addrType, err.Error()))
	}
	resp, err := clientx.Do(req)
	if err != nil {
		panic(fmt.Sprintf("Error updating %s address %s", addrType, err.Error()))
	}
	if resp.StatusCode != 200 {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(fmt.Sprintf("Error reading resp body %s", err.Error()))
		}
		panic(fmt.Sprintf("got non-200 status code on update. Exiting Status_Code=%d Msg=%s", resp.StatusCode, string(data)))
	}
}

func GetMasterAddress(controllerIP string, controllerPort int) (string, int) {
	log.Errorf("Querying for master address")
	resp, err := http.Get(fmt.Sprintf("http://%s:%d/address", controllerIP, controllerPort))
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

func GetState(controllerIP string, controllerPort int) string {
	resp, err := http.Get(fmt.Sprintf("http://%s:%d/state", controllerIP, controllerPort))
	if err != nil {
		log.Fatalf("Error getting state")
	}
	state, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Fatalf("Error reading state %s", err.Error())
	}
	return string(state)
}
