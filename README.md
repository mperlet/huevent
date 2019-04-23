# Huevent - Events for Philips-Hue buttons and sensors

> Simple program that terminates when a Philips Hue button is pressed.

## Example

```
./huevent -h
huevent - get events from buttons and sensors
Usage: ./huevent [OPTIONS] 
  -config string
    	path to config file (default "/home/mathias/.huevent/config.json")
  -debug
    	enable some debug output
  -exit
    	exit on event
  -pair
    	pair hue bridge
```

```
./huevent 
00:00:00:00:00:42:43:2f-f3      buttonevent	16
00:17:88:01:10:33:35:98-02-fc00	buttonevent	1000
00:17:88:01:03:29:57:55-02-0402	temperature	2134
00:17:88:01:03:29:57:55-02-0406	presence	false
00:17:88:01:03:29:57:55-02-0400	lightlevel	19888

```

## Supported Device and Event Types

| Device            | Event-Type (eventType) | Event-Data (triggerOn) | Description                                                   |
| ----------------- | ---------------------- | ---------------------- | ------------------------------------------------------------- |
| Hue Tap           | buttonevent            | 34                     | 1 Dot Key                                                     |
|                   | buttonevent            | 16                     | 2 Dot Key                                                     |
|                   | buttonevent            | 17                     | 3 Dot Key                                                     |
|                   | buttonevent            | 18                     | 4 Dot Key                                                     |
| Hue Dimmer Switch | buttonevent            | 1000                   | Hard press ON, followed by Event-Data 1002                    |
|                   | buttonevent            | 1001                   | Long press ON (send while hold the button)                    |
|                   | buttonevent            | 1003                   | Release ON (after Long press)                                 |
|                   | buttonevent            | 1002                   | Soft press ON                                                 |
|                   | buttonevent            | 2000                   | Hard press BRIGHTER, followed by Event-Data 2002              |
|                   | buttonevent            | 2001                   | Long press BRIGHTER (send while hold the button)              |
|                   | buttonevent            | 2003                   | Release BRIGHTER                                              |
|                   | buttonevent            | 2002                   | Soft press BRIGHTER                                           |
|                   | buttonevent            | 3000                   | Hard press DARKER, followed by Event-Data 3002                |
|                   | buttonevent            | 3001                   | Long press DARKER                                             |
|                   | buttonevent            | 3003                   | Release DARKER                                                |
|                   | buttonevent            | 3002                   | Soft press DARKER                                             |
|                   | buttonevent            | 4000                   | Hard press OFF, followed by Event-Data 4002                   |
|                   | buttonevent            | 4001                   | Long press OFF                                                |
|                   | buttonevent            | 4003                   | Release OFF                                                   |
|                   | buttonevent            | 4002                   | Soft press OFF                                                |
| Hue Motion Sensor | presence               | true                   | Motion detected                                               |
|                   | presence               | false                  | No Motion detected                                            |
|                   | temperature            | xxxx                   | Temperature in °C multiplied by 100 (2655 == 26.55°C)         |
|                   | lightlevel             | xxxxx                  | Lightlevel (0 ~ dark, 20000 ~ normal, 40000 ~ very bright)    |


## Configuration

Run `huevent -pair` to pair a local Hue Bridge.

### Configfile

`config`: Hue Bridge IP with Token

`hooks`: Execute commands on specific events

`deviceFilter`: Array(string) of deviceIds, ignore all other devices 


```
config:
  ip: 192.168....
  user: nEh...
hooks:
- deviceId: 00:17:88:01:10:33:35:98-02-fc00
  eventType: buttonevent
  triggerOn: 1002
  cmd: echo echo
- deviceId: 00:00:00:00:00:42:43:2f-f2
  eventType: buttonevent
  triggerOn: 18
  cmd: echo $HUEVENT_PAYLOAD NOOOOPE
- deviceId: 00:17:88:01:03:29:57:55-02-0406
  eventType: presence
  triggerOn: 'true'
  cmd: echo SOMEBODY BECOMES PRESENT
- deviceId: 00:17:88:01:03:29:57:55-02-0406
  eventType: presence
  triggerOn: 'false'
  cmd: echo SOMEBODY IS ABSENT
- deviceId: 00:17:88:01:03:29:57:55-02-0402
  eventType: temperature
  cmd: python -c 'import os; print(str(float(os.environ["HUEVENT_PAYLOAD"])/100.0) + "°C")'
deviceFilter: []
```

### Manual Pairing

```
# Press Hue Bridge Button before you ran these commands
BRIDGE_IP=$(curl https://www.meethue.com/api/nupnp -s | grep -E -o "([0-9]{1,3}[.]){3}[0-9]{1,3}")
USERNAME=$(curl -s -X POST -d'{"devicetype":"huevent"}' "http://$BRIDGE_IP/api" | grep -P -o '":"(.*)"' | cut -d '"' -f3)

echo "$BRIDGE_IP $USERNAME"

```

## Build

```
git clone https://github.com/mperlet/huevent.git && cd huevent
docker run --rm -v "$PWD":/huevent:Z -w /huevent -e GOOS=linux -e GOARCH=amd64 golang:1.12.4-stretch go build huevent.go
```

## Examples

* A Remote for Fritz Dect200 Power-Sockets
* Trigger <Insert-Your-Idea-Here>


## Licence

Copyright (c) 2019 Mathias Perlet

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
