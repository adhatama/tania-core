package nodered

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Tanibox/tania-core/config"
	"github.com/Tanibox/tania-core/src/assets/storage"
)

// We use JSON value in this initial template because it is easier.
// We don't need to make a struct for every possible node in node-red.
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
        "path": "{{.WebsocketConnectTo}}",
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

	resp, err := client.Get(*config.Config.NodeRedHost + ":" + *config.Config.NodeRedPort + "/flows")
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	existingBodyMap := make([]map[string]interface{}, 0)
	err = json.NewDecoder(resp.Body).Decode(&existingBodyMap)
	if err != nil {
		return err
	}

	// If no `id: main-flow-1` node, then just override the flow with initial template
	shouldOverride := true
	for _, v := range existingBodyMap {
		if v["id"] == "main-flow-1" {
			shouldOverride = false
			break
		}
	}

	var newFlow []byte
	if shouldOverride {
		// Generate new nodered flows

		tmpl, err := template.New("template").Parse(initialTemplate)
		if err != nil {
			return err
		}

		tmplData := struct {
			MQTTBrokerHost     string
			MQTTBrokerPort     string
			MQTTBrokerWsPort   string
			WebsocketConnectTo string
			NewNode            template.HTML
		}{
			*config.Config.MqttBrokerHost,
			*config.Config.MqttBrokerPort,
			*config.Config.MqttBrokerWsHost,
			*config.Config.WebsocketSensorConnectTo,
			"," + template.HTML(newTopicMarshalled),
		}

		var t bytes.Buffer
		err = tmpl.Execute(&t, tmplData)
		if err != nil {
			return err
		}
		fmt.Println(tmplData)

		newFlow = t.Bytes()
	} else {
		// Use existing nodered flows
		// If new node id is already exists, then just update it,
		// otherwise we just append that new node to the existing flows

		var newNodeMap map[string]interface{}
		err = json.Unmarshal(newTopicMarshalled, &newNodeMap)
		if err != nil {
			return err
		}

		isAlreadyExists := false
		for i, v := range existingBodyMap {
			if v["id"] == newNodeMap["id"] {
				isAlreadyExists = true
				existingBodyMap[i] = newNodeMap
			}
		}

		if !isAlreadyExists {
			existingBodyMap = append(existingBodyMap, newNodeMap)
		}

		newFlow, err = json.Marshal(existingBodyMap)
		if err != nil {
			return err
		}
	}

	req, err := http.NewRequest(
		"POST",
		*config.Config.NodeRedHost+":"+*config.Config.NodeRedPort+"/flows",
		bytes.NewBuffer(newFlow),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Node-RED-Deployment-Type", "nodes")

	postResp, err := client.Do(req)
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
