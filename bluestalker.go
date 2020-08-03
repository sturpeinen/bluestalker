package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"syscall"
)

type BluewalkerData struct {
	Device struct {
		Address string `json:"address"`
	} `json:"device"`
}

func mqtt_connect(host string, port int) (mqtt.Client, error) {
	opts := mqtt.NewClientOptions().
		SetAutoReconnect(true).
		AddBroker(fmt.Sprintf("tcp://%s:%d", host, port))

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return client, nil
}

func parse_address(data []byte) (string, error) {
	result := BluewalkerData{}
	err := json.Unmarshal(data, &result)
	if err != nil {
		return "", err
	}
	return result.Device.Address, nil
}

func main() {
	path := flag.String("unix", "/tmp/bluestalker.sock", "Path for the unix domain socket")
	mqtt_h := flag.String("host", "127.0.0.1", "MQTT host to connect to")
	mqtt_p := flag.Int("port", 1883, "MQTT port to connect to")
	mqtt_r := flag.Bool("retain", false, "MQTT message should be retained")
	topics_p := flag.String("topics", "", "Path for JSON file with topics (format {\"<address>\": \"/some/topic\", ..})")
	flag.Parse()

	var topics = make(map[string]string)
	if len(*topics_p) > 0 {
		data, err := ioutil.ReadFile(*topics_p)
		if err != nil {
			fmt.Println("Could not read topics:", err)
			return
		}
		err = json.Unmarshal(data, &topics)
		if err != nil {
			fmt.Println("Could not parse topics:", err)
			return
		}
	}

	if _, err := os.Stat(*path); !os.IsNotExist(err) {
		fmt.Println(*path, "already exists")
		return
	}

	sock, err := net.Listen("unix", *path)
	if err != nil {
		fmt.Println("Could not open Unix domain socket:", err)
		return
	}
	defer sock.Close()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		sock.Close()
		os.Exit(0)
	}()

	client, err := mqtt_connect(*mqtt_h, *mqtt_p)
	if err != nil {
		fmt.Println("Could not connect to MQTT broker:", err)
		return
	}

	for {
		conn, err := sock.Accept()
		if err != nil {
			continue
		}

		reader := bufio.NewReader(conn)
		for {
			line, _, err := reader.ReadLine()
			if err == io.EOF {
				break
			}
			address, err := parse_address(line)
			if len(address) <= 0 || err != nil {
				continue
			}

			topic := fmt.Sprintf("bluewalker/%s", address)
			if t, ok := topics[address]; ok {
				topic = t
			}
			client.Publish(topic, 0, *mqtt_r, string(line))
			fmt.Println(topic, string(line))
		}
	}
}
