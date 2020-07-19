package main

import (
    "bufio"
    "encoding/json"
    "flag"
    "fmt"
    "io"
    "net"
    "os"
    "os/signal"
    "syscall"
    mqtt "github.com/eclipse/paho.mqtt.golang"
)

type BluewalkerData struct {
    Device struct { Address string `json:"address"`} `json:"device"`
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
    flag.Parse()

    if _, err := os.Stat(*path); !os.IsNotExist(err) {
		fmt.Println(*path, "already exists")
        return
	}

    sock, err := net.Listen("unix", *path);
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
            client.Publish(topic, 0, *mqtt_r, string(line))
            fmt.Println(topic, string(line))
        }
    }
}
