# bluestalker

```
$ bluestalker -help
Usage of bluestalker:
  -host string
       MQTT host to connect to (default "127.0.0.1")
  -port int
       MQTT port to connect to (default 1883)
  -retain
       MQTT message should be retained
  -topics string
       Path for JSON file with topics (format {"<address>": "/some/topic", ..})
  -unix string
       Path for the unix domain socket (default "/tmp/bluestalker.sock")
```

## Usage
Run **bluestalker**.
```
$ bluestalker -host 192.0.2.42 -unix /tmp/bs.sock
...
```

Run and connect [Bluewalker](https://gitlab.com/jtaimisto/bluewalker/) to previously specified unix domain socket. In this example Bluewalker is configured to scan [RuuviTags](https://ruuvi.com/ruuvitag-specs/).
```
# bluewalker -device hci0 -duration -1 -json -unix /tmp/bs.sock -ruuvi
...
```

Verify from MQTT broker that everything is working. Default topic for messages is ```bluewalker/<address>```. Topic can be replaced with JSON topic mapping, see below.
```
$ mosquitto_sub -h 192.0.2.42 -t "bluewalker/#" -v
bluewalker/de:ad:00:03:13:37 {"device":{"address":"de:ad:00:03:13:37","type":"LE Random"},"rssi":-71,"sensors":{"humidity":44,"temperature":66.6,"pressure":101445,"accelerationX":-0.868,"accelerationY":-0.508,"accelerationZ":-0.036,"voltage":2827,"txpower":31,"movementCount":255,"sequence":65535}}
...
```

Optionally, topics flag (```-topics /path/to/file.json```) can be used with **bluestalker** to provide MQTT topic mapping for device addresses. JSON example:
```
{
    "03:13:37:00:be:ef": "hell/cold",
    "de:ad:00:03:13:37": "sauna/hot"
}
```

## Build
Install [Go](https://golang.org/doc/install) and run ```go build```.

```
$ cd bluestalker
$ go build
$ file bluestalker
bluestalker: ELF 64-bit LSB executable, x86-64, version 1 (SYSV), dynamically linked, interpreter /lib64/ld-linux-x86-64.so.2, ...
```

Cross compile for Raspberry Pi.
```
$ cd bluestalker
$ GOOS=linux GOARCH=arm GOARM=5 go build
$ file bluestalker
bluestalker: ELF 32-bit LSB executable, ARM, EABI5 version 1 (SYSV), statically linked, ...
```
