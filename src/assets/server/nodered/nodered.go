package nodered

import (
	"bytes"
	"encoding/json"
	"errors"
	"html/template"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Tanibox/tania-core/src/assets/storage"
)

const initialTemplate = `
[
    {
        "id": "main-flow-1",
        "type": "tab",
        "label": "Main Flow 1",
        "disabled": false,
        "info": ""
    },
    {
        "id": "main-flow-1-mqtt-broker",
        "type": "mqtt-broker",
        "z": "main-flow-1",
        "name": "mqtt-broker",
        "broker": "{{.MQTTBrokerHost}}",
        "port": "{{.MQTTBrokerPort}}",
        "clientid": "",
        "usetls": false,
        "compatmode": true,
        "keepalive": "60",
        "cleansession": true,
        "birthTopic": "",
        "birthQos": "0",
        "birthPayload": "",
        "closeTopic": "",
        "closeQos": "0",
        "closePayload": "",
        "willTopic": "",
        "willQos": "0",
        "willPayload": ""
    },
    {
        "id": "main-flow-1-websocket-client",
        "type": "websocket-client",
        "z": "main-flow-1",
        "path": "{{.WebsocketClientPath}}",
        "tls": "",
        "wholemsg": "false"
    },
    {
        "id": "main-flow-1-mosca",
        "type": "mosca in",
        "z": "main-flow-1",
        "mqtt_port": {{.MQTTBrokerPort}},
        "mqtt_ws_port": {{.MQTTBrokerWsPort}},
        "name": "",
        "username": "",
        "password": "",
        "dburl": "",
        "x": 149.5,
        "y": 109,
        "wires": [
            [
                "main-flow-1-mosca-debug"
            ]
        ]
    },
    {
        "id": "main-flow-1-mosca-debug",
        "type": "debug",
        "z": "main-flow-1",
        "name": "",
        "active": true,
        "tosidebar": true,
        "console": false,
        "tostatus": false,
        "complete": "payload",
        "x": 398.5,
        "y": 109,
        "wires": []
    },
    {
        "id": "main-flow-1-mqtt-my-second-topic",
        "type": "mqtt in",
        "z": "main-flow-1",
        "name": "",
        "topic": "my-second-topic",
        "qos": "0",
        "broker": "main-flow-1-mqtt-broker",
        "x": 150,
        "y": 175,
        "wires": [
            [
                "main-flow-1-websocket-out"
            ]
        ]
	},
	{
        "id": "main-flow-1-mqtt-my-third-topic",
        "type": "mqtt in",
        "z": "main-flow-1",
        "name": "",
        "topic": "my-third-topic",
        "qos": "0",
        "broker": "main-flow-1-mqtt-broker",
        "x": 150,
        "y": 225,
        "wires": [
            [
                "main-flow-1-websocket-out"
            ]
        ]
    },
    {
        "id": "main-flow-1-websocket-out",
        "type": "websocket out",
        "z": "main-flow-1",
        "name": "websocket client",
        "server": "",
        "client": "main-flow-1-websocket-client",
        "x": 500,
        "y": 175,
        "wires": []
    }{{.NewNode}}
]
`

func Update(device *storage.DeviceRead) error {
	newTopicNode := struct {
		ID     string     `json:"id"`
		Type   string     `json:"type"`
		Z      string     `json:"z"`
		Name   string     `json:"name"`
		Topic  string     `json:"topic"`
		Qos    string     `json:"qos"`
		Broker string     `json:"broker"`
		X      int        `json:"x"`
		Y      int        `json:"y"`
		Wires  [][]string `json:"wires"`
	}{
		ID:     device.DeviceID,
		Type:   "mqtt in",
		Z:      "main-flow-1",
		Topic:  device.TopicName,
		Qos:    "0",
		Broker: "main-flow-1-mqtt-broker",
		X:      150,
		Y:      225,
	}
	newTopicNode.Wires = append(newTopicNode.Wires, []string{"main-flow-1-websocket-out"})

	newTopicMarshalled, err := json.Marshal(newTopicNode)
	if err != nil {
		return err
	}

	var client = &http.Client{
		Timeout: time.Second * 10,
	}

	resp, err := client.Get("http://localhost:1880/flows")
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	bodyJSON := make([]map[string]interface{}, 0)
	err = json.NewDecoder(resp.Body).Decode(&bodyJSON)
	if err != nil {
		return err
	}

	var newFlow []byte
	if len(bodyJSON) == 0 {
		// Generate new nodered flows

		tmpl, err := template.New("template").Parse(initialTemplate)
		if err != nil {
			return err
		}

		tmplData := struct {
			MQTTBrokerHost      string
			MQTTBrokerPort      string
			MQTTBrokerWsPort    string
			WebsocketClientPath string
			NewNode             template.HTML
		}{
			"localhost",
			"1883",
			"8080",
			"ws://localhost:8080/ws/test/heru",
			"," + template.HTML(newTopicMarshalled),
		}

		var t bytes.Buffer
		err = tmpl.Execute(&t, tmplData)
		if err != nil {
			return err
		}

		newFlow = t.Bytes()
	} else {
		// Use existing nodered flows

		var newNodeMap map[string]interface{}
		err = json.Unmarshal(newTopicMarshalled, &newNodeMap)
		if err != nil {
			return err
		}

		bodyJSON = append(bodyJSON, newNodeMap)

		newFlow, err = json.Marshal(bodyJSON)
		if err != nil {
			return err
		}
	}

	postResp, err := client.Post("http://localhost:1880/flows", "application/json", bytes.NewBuffer(newFlow))
	if err != nil {
		return err
	}
	defer postResp.Body.Close()

	if postResp.StatusCode != http.StatusNoContent {
		postRespBody, err := ioutil.ReadAll(postResp.Body)
		if err != nil {
			return err
		}

		return errors.New(string(postRespBody))
	}

	return nil
}
