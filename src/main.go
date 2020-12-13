package main

//import "time"
import "io"
import "bufio"
import "fmt"
import "os"
import "strings"
import "os/exec"
import MQTT "github.com/eclipse/paho.mqtt.golang"

type Message struct {
	Topic   string
	Payload string
}

var messages = make(chan Message)
var commands = make(chan string)

func defaultHandler(client MQTT.Client, msg MQTT.Message) {
	message := Message{msg.Topic(), string(msg.Payload())}
	fmt.Printf("RECEIVED TOPIC: %s MESSAGE: %s\n", message.Topic, message.Payload)
	commands <- fmt.Sprintf("%s\n", message.Payload)
}

func setupMqtt() {
	broker := "tcp://iot.labs:1883"
	id := "MQTT_to_CEC"
	topic := "KODI_ON"

	opts := MQTT.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(id)
	opts.SetDefaultPublishHandler(defaultHandler)
	client := MQTT.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	fmt.Println("Connected")

	if token := client.Subscribe(topic, 0, nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	fmt.Println("Subscribed")

	//client.Disconnect(250)
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
			_, err = stdin.Write([]byte(command))
			if err != nil {
				panic(err)
			}
			fmt.Printf("Wrote <%s>\n", strings.TrimSpace(command))
		}
	}()

	go func() {
		for true {
			answer, err := reader.ReadString('\n')
			if err == io.EOF {
				//close(answers)
				fmt.Printf("Blowing up")
				return
			}
			// FIXME if FD is closed i die
			if err != nil {
				panic(err)
			}
			answer = strings.TrimSpace(answer)
			fmt.Printf("Got back: <%s>\n", answer)
			//answers <- answer
			// FIXME read from here
		}
	}()

	c.Wait()
}
