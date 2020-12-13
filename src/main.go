package main

import "bufio"
import "fmt"
import "os"
import "strings"
import "os/exec"
import MQTT "github.com/eclipse/paho.mqtt.golang"

var commands = make(chan string)

func defaultHandler(client MQTT.Client, msg MQTT.Message) {
	fmt.Printf("< [%s]: %s\n", msg.Topic(), string(msg.Payload()))
	commands <- string(msg.Payload())
}

func setupMqtt() {
	broker := getEnv("MQTT_BROKER", "tcp://iot.labs:1883")
	topic := getEnv("MQTT_TOPIC", "KODI_ON")
	id := "MQTT_to_CEC"

	fmt.Printf("Connecting to %s as %s\n", broker, id)
	opts := MQTT.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(id)
	opts.SetDefaultPublishHandler(defaultHandler)
	client := MQTT.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	fmt.Println("Connected")
	fmt.Printf("Subscribing to %s\n", topic)
	if token := client.Subscribe(topic, 0, nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	fmt.Println("Subscribed")
}

func main() {
	setupMqtt()
	c := exec.Command("cec-client", "-d", "1")
	stdin, err := c.StdinPipe()
	if err != nil {
		panic(err)
	}
	stdout, err := c.StdoutPipe()
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(stdout)
	err = c.Start()
	if err != nil {
		panic(err)
	}

	go func() {
		for command := range commands {
			_, err = stdin.Write([]byte(fmt.Sprintf("%s\n", command)))
			if err != nil {
				panic(err)
			}
			fmt.Printf("Wrote <%s>\n", strings.TrimSpace(command))
		}
	}()

	go func() {
		for true {
			answer, err := reader.ReadString('\n')
			if err != nil {
				panic(err)
			}
			answer = strings.TrimSpace(answer)
			fmt.Printf("Got back: <%s>\n", answer)
		}
	}()

	c.Wait()
}
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
