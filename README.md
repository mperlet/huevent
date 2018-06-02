# Huevent - Events for Philips Hue Buttons/Switches

> Simple program that terminates when a Philips Hue button is pressed.

## Pattern

Huevent blocks as long as a button is pressed. The Button ID is written to stdout. 
Other tools like curl can then be used to trigger an event to another level.

```
# Bash Example: make GET request to a local http service on button press
while true; do curl "localhost:8000/button/$(huevent)"; done
```

## Install

TODO

## Configuration

Huevent need two things to work as expected. 

* Bridge URI, which contains the Bridge IP/Hostname and a registered User
  * Like: `http://192.168.178.38/api/f1rL3FaNzMuMp1tZf1rL3FaNzMuMp1tZ/sensors`
* Sensor-ID, the unique identifier of your button, make request to the /sensors endpoint to find your id
  * Like: `00:00:00:00:00:ca:ff:ee-f2`

To find your bridge and create a user read the [Getting Started Guide](https://developers.meethue.com/documentation/getting-started)

You can also try this bash-script: 
```
# Press Hue Bridge Button before you ran these commands
BRIDGE_IP=$(curl https://www.meethue.com/api/nupnp -s | grep -E -o "([0-9]{1,3}[.]){3}[0-9]{1,3}")
USERNAME=$(curl -s -X POST -d'{"devicetype":"huevent"}' "http://$BRIDGE_IP/api" | grep -P -o '":"(.*)"' | cut -d '"' -f3)

echo "export HUEVENT_URI=http://$BRIDGE_IP/api/$USERNAME/sensors"
echo "export HUEVENT_ID="<YOUR SENSOR ID>"
```

## Examples

* A Remote for Fritz Dect200 Power-Sockets
* Trigger <Insert-Your-Idea-Here>

### Planned Features 

* Commandline Arguments (version, help)
* Multi Button
* Temperature/Light Sensor Value Change Events


## Licence

Copyright (c) 2018 Mathias Perlet

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
