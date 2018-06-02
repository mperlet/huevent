package main

import "fmt"
import "net/http"
import "os"
import "io/ioutil"
import "encoding/json"
import "time"

type config struct {
	uri            string
	deviceId       string
	buttonStateMap *map[int]string
	responseMap    *map[string]interface{}
	pollTimeMs     time.Duration
	logHTTPError   bool
}

func main() {

	var conf = makeConfig()

	for {
		poll(&conf)
		time.Sleep(conf.pollTimeMs * time.Millisecond)
	}

}

func makeConfig() config {
	buttonStateMap := make(map[int]string)
	responseMap := make(map[string]interface{})
	return config{
		uri:            os.Getenv("HUEVENT_URI"),
		deviceId:       os.Getenv("HUEVENT_ID"),
		buttonStateMap: &buttonStateMap,
		responseMap:    &responseMap,
		pollTimeMs:     333,
		logHTTPError:   true}
}

func poll(conf *config) {
	resp, err := http.Get(conf.uri)

	if err != nil {
		if conf.logHTTPError {
			fmt.Fprintln(os.Stderr, err)
			conf.logHTTPError = false
		}
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		json.Unmarshal(body, conf.responseMap)
		parseJSONMap(conf.responseMap, conf)
	} else {
		fmt.Fprintln(os.Stderr, err)
	}

	if !conf.logHTTPError {
		conf.logHTTPError = true
	}
}

func exit(keyCode int) {
	fmt.Println(keyCode)
	os.Exit(0)
}

func updateButtonMap(update map[string]interface{}, conf *config) {

	var key = int(update["buttonevent"].(float64))
	var value = update["lastupdated"].(string)

	// check for known button
	if val, ok := (*conf.buttonStateMap)[key]; ok {
		// check if button pressed
		if val != value {
			exit(key)
		}
	} else {
		// unknown button, check for initial run
		if len((*conf.buttonStateMap)) != 0 {
			exit(key)
		}
	}

	(*conf.buttonStateMap)[key] = value
}

func parseJSONMap(jsonAsMap *map[string]interface{}, conf *config) {
	for _, v := range *jsonAsMap {
		if subObject, ok := v.(map[string]interface{}); ok {
			if deviceId, ok := subObject["uniqueid"]; ok {
				if stateObject, ok := subObject["state"]; ok && deviceId == conf.deviceId {
					updateButtonMap(stateObject.(map[string]interface{}), conf)
				}
			}
			parseJSONMap(&subObject, conf)
		}
	}
}
