package main

import "fmt"
import "net/http"
import "os"
import "io/ioutil"
import "encoding/json"
import "time"

type config struct {
	uri          string
	deviceIds    []string
	hasFilter    bool
	stateMap     *map[string]map[string]string
	responseMap  *map[string]interface{}
	pollTimeMs   time.Duration
	logHTTPError bool
}

func main() {
	var conf = makeConfig(os.Args[1:])

	for {
		poll(&conf)
		time.Sleep(conf.pollTimeMs * time.Millisecond)
	}

}

func makeConfig(deviceIds []string) config {
	stateMap := make(map[string]map[string]string)
	var hasFilter = false

	for _, deviceID := range deviceIds {
		stateMap[deviceID] = make(map[string]string)
		hasFilter = true
	}
	responseMap := make(map[string]interface{})
	return config{
		uri:          os.Getenv("HUEVENT_URI"),
		deviceIds:    deviceIds,
		stateMap:     &stateMap,
		responseMap:  &responseMap,
		hasFilter:    hasFilter,
		pollTimeMs:   333,
		logHTTPError: true}
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

func exit(device string, eventType string, keyCode string) {
	fmt.Printf("%s\t%s\t%s\n", device, eventType, keyCode)
	os.Exit(0)
}

func updateButtonMap(update map[string]interface{}, conf *config, device1 interface{}) {
	var device = device1.(string)

	var btnStateMap = (*conf.stateMap)[device]

	var key = "unknown"
	var eventType = "unknown"

	if update["buttonevent"] != nil {
		key = fmt.Sprintf("%v", update["buttonevent"].(float64))
		eventType = "buttonevent"
	}

	if update["presence"] != nil {
		key = fmt.Sprintf("%v", update["presence"].(bool))
		eventType = "presence"
	}

	if update["lightlevel"] != nil {
		key = fmt.Sprintf("%v", update["lightlevel"].(float64))
		eventType = "lightlevel"
	}

	if update["temperature"] != nil {
		key = fmt.Sprintf("%v", update["temperature"].(float64))
		eventType = "temperature"
	}

	if eventType == "unknown" {
		return
	}

	var value = update["lastupdated"].(string)
	// check for known button
	if val, ok := btnStateMap[key]; ok {
		// check if button pressed
		if val != value {
			exit(device, eventType, key)
		}
	} else {
		// unknown button, check for initial run
		if len(btnStateMap) != 0 {
			exit(device, eventType, key)
		}
	}

	btnStateMap[key] = value
}

func parseJSONMap(jsonAsMap *map[string]interface{}, conf *config) {
	for _, v := range *jsonAsMap {
		if subObject, ok := v.(map[string]interface{}); ok {
			if deviceId, ok := subObject["uniqueid"]; ok {
				addNewSensorToStateMap(deviceId, conf)
				if stateObject, ok := subObject["state"]; ok && hasKey(deviceId, conf.stateMap) {
					updateButtonMap(stateObject.(map[string]interface{}), conf, deviceId)
				}
			}
			parseJSONMap(&subObject, conf)
		}
	}
}

func addNewSensorToStateMap(deviceId interface{}, conf *config) {
	if conf.hasFilter {
		// do not add a sensor if the argument filter is enabled
		return
	}

	if _, ok := (*conf.stateMap)[deviceId.(string)]; ok {
		// do not add the sensor if the sensor is still added
		return
	}
	// allocate a new state map, add device to stateMap
	(*conf.stateMap)[deviceId.(string)] = make(map[string]string)
}

func hasKey(a interface{}, map1 *map[string]map[string]string) bool {
	_, ok := (*map1)[a.(string)]
	return ok
}
