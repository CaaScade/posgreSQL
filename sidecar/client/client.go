package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/caascade/posgreSQL/posgresql/resource"
)

func GetSlaveAddress(controllerIP string, controllerPort int) (string, int) {
	for {
		//resp, err := http.Get(fmt.Sprintf("http://%s:%d/address", controllerIP, controllerPort))
		resp, err := http.Get(fmt.Sprintf("http://%s/address", controllerIP))
		if err == nil {
			var addresses resource.Addresses
			body, err := ioutil.ReadAll(resp.Body)
			defer resp.Body.Close()
			if err != nil {
				time.Sleep(2 * time.Second)
				continue
			}
			err = json.Unmarshal(body, &addresses)
			if err == nil {
				if addresses.SlaveIP != "" && addresses.SlavePort != 0 {
					return addresses.SlaveIP, addresses.SlavePort
				}
			}
		}
		time.Sleep(2 * time.Second)
	}
	return "", 0
}

func GetMasterAddress(controllerIP string, controllerPort int) (string, int) {
	for {
		//resp, err := http.Get(fmt.Sprintf("http://%s:%d/address", controllerIP, controllerPort))
		resp, err := http.Get(fmt.Sprintf("http://%s/address", controllerIP))
		if err == nil {
			var addresses resource.Addresses
			body, err := ioutil.ReadAll(resp.Body)
			defer resp.Body.Close()
			if err != nil {
				time.Sleep(2 * time.Second)
				continue
			}
			err = json.Unmarshal(body, &addresses)
			if err == nil {
				if addresses.MasterIP != "" && addresses.MasterPort != 0 {
					return addresses.MasterIP, addresses.MasterPort
				}
			}
		}
		time.Sleep(2 * time.Second)
	}
	return "", 0
}
