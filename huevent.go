package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"time"
)

type config struct {
	uri          string
	deviceIds    []string
	hasFilter    bool
	stateMap     *map[string]map[string]string
	responseMap  *map[string]interface{}
	pollTimeMs   time.Duration
	logHTTPError bool
	shouldExit   bool
	hooks        *[]Hook
	DEBUG        bool
	printSensors bool
}

type HueventConfig struct {
	Config struct {
		Ip   string `yaml:"ip"`
		User string `yaml:"user"`
		Rate int `yaml:"pollingRateMs,omitempty"`
	} `yaml:"config"`
	Hooks        []Hook   `yaml:"hooks"`
	DeviceFilter []string `yaml:"deviceFilter"`
}

type Hook struct {
	DeviceID  string `yaml:"deviceId"`
	EventType string `yaml:"eventType"`
	TriggerOn string `yaml:"triggerOn,omitempty"`
	Cmd       string `yaml:"cmd"`
}

type hueBridgeResponse struct {
	ID                string `json:"id"`
	InternalIpaddress string `json:"internalipaddress"`
}

type hueBridgePairResponse struct {
	Error struct {
		Type        string
		Address     string
		Description string
	}
	Success struct {
		Username string
	}
}

// DEBUG huevent debug flag
var DEBUG = false

func main() {
	var conf = makeConfig()

	if conf.DEBUG {
		fmt.Printf("current configuration: %#v\n", conf)
	}

	for {
		poll(&conf)
		time.Sleep(conf.pollTimeMs * time.Millisecond)
	}

}

func myUsage() {
	fmt.Printf("huevent - get events from buttons and sensors\n")
	fmt.Printf("Usage: %s [OPTIONS] \n", os.Args[0])
	flag.PrintDefaults()
}

func pairBridge(configpath string) {

	if DEBUG {
		fmt.Printf("pair bridge, ask https://www.meethue.com/api/nupnp\n")
	}

	resp, err := http.Get("https://www.meethue.com/api/nupnp")

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}

	if DEBUG {
		fmt.Printf("response from https://www.meethue.com/api/nupnp %s\n", body)
	}

	var hueBridges = []hueBridgeResponse{}
	unmarshalErr := json.Unmarshal(body, &hueBridges)

	if unmarshalErr != nil {
		panic(unmarshalErr)
	}

	if len(hueBridges) == 0 {
		fmt.Printf("no bridges found\n")
		os.Exit(1)
	}

	values := map[string]string{"devicetype": "huevent"}

	jsonValue, _ := json.Marshal(values)

	var uri = fmt.Sprintf("http://%s/api", hueBridges[0].InternalIpaddress)

	resp, err = http.Post(uri, "application/json", bytes.NewBuffer(jsonValue))

	//noinspection ALL
	body, err = ioutil.ReadAll(resp.Body)

	if DEBUG {
		fmt.Printf("response pairing %s:  %s\n", hueBridges[0].InternalIpaddress, body)
	}

	var hueResponse []hueBridgePairResponse
	_ = json.Unmarshal(body, &hueResponse)

	if hueResponse[0].Success.Username == "" {
		fmt.Printf("Error while pairing with %s: %s \n", hueBridges[0].InternalIpaddress, hueResponse[0].Error.Description)
		os.Exit(1)
	}

	var pairingSuccess = map[string]map[string]string{}
	pairingSuccess["config"] = map[string]string{}
	pairingSuccess["config"]["ip"] = hueBridges[0].InternalIpaddress
	pairingSuccess["config"]["user"] = hueResponse[0].Success.Username

	var currentConfig = readConfig(configpath)
	currentConfig.Config.Ip = hueBridges[0].InternalIpaddress
	currentConfig.Config.User = hueResponse[0].Success.Username
	writeConfig(currentConfig, configpath)

	os.Exit(0)
}

func makeConfig() config {

	exitOnEvent := flag.Bool("exit", false, "exit on event")
	hueventConfigPath := flag.String("config", configPath(), "path to config file")
	printSensors := flag.Bool("sensors", false, "print sensors response and exit")
	debug := flag.Bool("debug", false, "enable some debug output")

	pair := flag.Bool("pair", false, "pair hue bridge")

	flag.Usage = myUsage
	flag.Parse()

	DEBUG = *debug

	if *pair {
		pairBridge(*hueventConfigPath)
	}

	var hueventConfig = readConfig(*hueventConfigPath)

	if *debug {
		fmt.Printf("Read Config %+s\n", hueventConfig)
	}

	if hueventConfig.Config.Ip == "" || hueventConfig.Config.User == "" {
		fmt.Printf("Error: no Hue-Bridge configured at %s, run huevent with -pair to configure\n", *hueventConfigPath)
		os.Exit(1)
	}
	
	var pollTimeMs = time.Duration(333)

	if hueventConfig.Config.Rate > 0 {
		pollTimeMs = time.Duration(hueventConfig.Config.Rate)
	}
	
	if *debug {
		fmt.Printf("Polling-Rate: %s\n", pollTimeMs * time.Millisecond)
	}

	stateMap := make(map[string]map[string]string)
	var hasFilter = false
	var deviceIds = hueventConfig.DeviceFilter

	for _, deviceID := range deviceIds {
		stateMap[deviceID] = make(map[string]string)
		hasFilter = true
	}

	responseMap := make(map[string]interface{})
	return config{
		uri:          fmt.Sprintf("http://%s/api/%s/sensors", hueventConfig.Config.Ip, hueventConfig.Config.User),
		deviceIds:    deviceIds,
		stateMap:     &stateMap,
		responseMap:  &responseMap,
		hasFilter:    hasFilter,
		pollTimeMs:   pollTimeMs,
		logHTTPError: true,
		hooks:        &hueventConfig.Hooks,
		shouldExit:   *exitOnEvent,
		DEBUG:        *debug,
		printSensors: *printSensors}
}

func poll(conf *config) {
	resp, err := http.Get(conf.uri)

	if err != nil {
		if conf.logHTTPError {
			_, _ = fmt.Fprintln(os.Stderr, err)
			conf.logHTTPError = false
		}
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		_ = json.Unmarshal(body, conf.responseMap)
		parseJSONMap(conf.responseMap, conf)
	} else {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}

	if !conf.logHTTPError {
		conf.logHTTPError = true
	}
	
	if conf.printSensors {
		var prettyJSON bytes.Buffer
		_ = json.Indent(&prettyJSON, body, "", "\t")
		fmt.Printf("%s\n", prettyJSON.String())
		os.Exit(0)
	}
}

func exit(device string, eventType string, triggerOn string, conf *config) {
	fmt.Printf("%s\t%s\t%s\n", device, eventType, triggerOn)

	for _, hook := range *conf.hooks {

		if hook.DeviceID != device || hook.EventType != eventType {
			continue
		}

		if hook.TriggerOn == "" || hook.TriggerOn == triggerOn {
			//noinspection ALL
			go executeCommand(hook.Cmd, device, eventType, triggerOn)
		}

	}

	if conf.shouldExit {
		os.Exit(0)
	}
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
			exit(device, eventType, key, conf)
		}
	} else {
		// unknown button, check for initial run
		if len(btnStateMap) != 0 {
			exit(device, eventType, key, conf)
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

func executeCommand(cmdString string, deviceId string, eventType string, payload string) error {
	cmd := exec.Command("/bin/sh", "-c", cmdString)

	extraEnv := []string{
		fmt.Sprintf("HUEVENT_DEVICE_ID=%s", deviceId),
		fmt.Sprintf("HUEVENT_EVENT_TYPE=%s", eventType),
		fmt.Sprintf("HUEVENT_PAYLOAD=%s", payload)}

	cmd.Env = append(os.Environ(), extraEnv...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func configPath() string {
	var configDirectory = path.Join(os.Getenv("HOME"), ".huevent")
	return path.Join(configDirectory, "huevent.yml")
}

func readConfig(filepath string) HueventConfig {
	var hueventConfig = HueventConfig{}

	content, readErr := ioutil.ReadFile(filepath)

	if readErr != nil {
		writeConfig(HueventConfig{}, filepath)
	}

	unmarshalErr := yaml.Unmarshal(content, &hueventConfig)
	if unmarshalErr != nil {
		fmt.Printf("Can't parse config file (%s), delete it and create a new with -pair argument", filepath)
	}
	return hueventConfig
}

func writeConfig(config HueventConfig, filepath string) {

	var a, err = yaml.Marshal(&config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	if DEBUG {
		fmt.Printf("Write Config %s \n %s \n", filepath, a)
	}

	if !PathExists(filepath) {
		var configDir = path.Dir(filepath)
		dirErr := os.MkdirAll(configDir, 0755)
		if dirErr != nil {
			panic(dirErr)
		}
	}

	writeErr := ioutil.WriteFile(filepath, a, 0644)
	if writeErr != nil {
		panic(writeErr)
	}
}
