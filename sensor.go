package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
)

const SERVER_URL = "http://192.168.1.90:5000"

type ComponentType int

const (
	LED_ACTUATOR = -1

	LIGHT_SENSOR  = 1
	DHT11_SENSOR  = 2
	MOTION_SENSOR = 3
)

func (ct ComponentType) GetName() string {
	switch ct {
	case LED_ACTUATOR:
		return "LED"
	case LIGHT_SENSOR:
		return "Light"
	case DHT11_SENSOR:
		return "DHT11"
	case MOTION_SENSOR:
		return "Motion"
	default:
		return "Unknown"
	}
}

// LEDActuatorData | LightSensorData | TemperatureSensorData | HumiditySensorData | MotionSensorData
type ComponentData interface {
	GetVal() interface{}
	SetVal(Val interface{})
}

type Component struct {
	Id   uint64
	Name string
	Room string
	Type ComponentType
	Data ComponentData
}

type LEDActuatorData struct{ Val bool }

func (l LEDActuatorData) GetVal() interface{} {
	return l.Val
}

func (l *LEDActuatorData) SetVal(Val interface{}) {
	l.Val = Val.(bool)
}

type LightSensorData struct{ Val uint8 }

func (l LightSensorData) GetVal() interface{} {
	return l.Val
}

func (l *LightSensorData) SetVal(Val interface{}) {
	l.Val = Val.(uint8)
}

type DHT11SensorDataVal struct {
	Temperature float32
	Humidity    float32
}

type DHT11SensorData struct {
	Val DHT11SensorDataVal
}

func (d DHT11SensorData) GetVal() interface{} {
	return d.Val
}

func (d *DHT11SensorData) SetVal(Val interface{}) {
	d.Val = Val.(DHT11SensorDataVal)
}

type MotionSensorData struct{ Val bool }

func (m MotionSensorData) GetVal() interface{} {
	return m.Val
}

func (m *MotionSensorData) SetVal(Val interface{}) {
	m.Val = Val.(bool)
}

type House map[string][]Component

func PopulateHouseWithComponents(house *House, components []Component) {
	var component Component
	for i := 0; i < len(components); i++ {
		component = components[i]

		components_in_room, room_exists := (*house)[component.Room]

		if !room_exists {
			(*house)[component.Room] = []Component{component}
			continue
		}

		//// house[component.Room] = append(components_in_room, component)

		component_is_in_room := false
		for j := 0; j < len(components_in_room); j++ {
			if components_in_room[j].Id == component.Id {
				components_in_room[j].Name = component.Name
				components_in_room[j].Data.SetVal(component.Data.GetVal())
				component_is_in_room = true
			}
		}

		if !component_is_in_room {
			(*house)[component.Room] = append((*house)[component.Room], component)
		}

	}
}

func (house House) GetRooms() []string {
	rooms := make([]string, 0, len(house))
	for r := range house {
		rooms = append(rooms, r)
	}
	sort.Strings(rooms)

	return rooms
}

// ************************ REQUESTS LOGIC ************************

// The JsonComponent is used for parsing the json recieved from the server.
// It is very similar to the Sensor struct and the only reason why it exists is because
// of the Data and Val attributes, as they are different for each of the component
// types. The server also has all of the attributes in lower case and instead of
// changing the API based on the frontend (Go only exports fields with capital letters,
// but python doesn't care about the capital letters), I decided to use the struct tags
type JsonComponent struct {
	Id   uint64        `json:"id"`
	Name string        `json:"name"`
	Room string        `json:"room"`
	Type ComponentType `json:"type"`
	Val  interface{}   `json:"val"`
}

func (json_component JsonComponent) ToComponent() Component {
	return Component{
		Id:   json_component.Id,
		Name: json_component.Name,
		Room: json_component.Room,
		Type: json_component.Type,
		Data: func() ComponentData {
			switch json_component.Type {
			case LED_ACTUATOR:
				return &LEDActuatorData{
					Val: json_component.Val.(bool),
				}
			case LIGHT_SENSOR:
				return &LightSensorData{
					// The json library casts every numeric item to a float64 apparently
					Val: uint8(json_component.Val.(float64)),
				}
			case DHT11_SENSOR:
				return &DHT11SensorData{
					Val: DHT11SensorDataVal{
						// The json library also casts arrays to []interface{} also...
						Temperature: float32(json_component.Val.([]interface{})[0].(float64)),
						Humidity:    float32(json_component.Val.([]interface{})[1].(float64)),
					},
				}
			case MOTION_SENSOR:
				return &MotionSensorData{
					Val: json_component.Val.(bool),
				}
			default:
				panic("Unknown component type")
			}
		}(),
	}
}

func GetComponents(components *[]Component) {
	resp, err := http.Get(SERVER_URL + "/get_components")
	if err != nil {
		fmt.Printf("Get Request Error: %v\n", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Reading Response Error: %v\n", err)
	}

	var components_json []JsonComponent

	err = json.Unmarshal(body, &components_json)
	if err != nil {
		fmt.Printf("Json Unmarshaling Error: %v\n", err)
		fmt.Printf("body: %v\n", string(body))
	}

	for i, json_component := range components_json {
		if i == len(*components) {
			*components = append(*components, json_component.ToComponent())
			continue
		}

		component := &(*components)[i]

		// If a component does not have same id as json_component's id, it must have been deleted
		// ! However this logic is not the best and requires the components to only be added at the end !
		// TODO: Change this to make it better, or more logically sound (maybe it is but I confused myself?)
		if component.Id != json_component.Id {
			// Removes the component by creating a new slice without it
			*components = append((*components)[:i], (*components)[i+1:]...)
			continue
		}

		// It is possible for a component's name, room and value to change, so this will track any changes
		component.Name = json_component.Name
		component.Room = json_component.Room
		component.Data.SetVal(json_component.ToComponent().Data.GetVal())
	}
}

type UpdateActuatorValueBody struct {
	Id  uint64      `json:"id"`
	Val interface{} `json:"val"`
}

func UpdateActuatorValue(component_id uint64, new_val interface{}) {
	body, err := json.Marshal(
		UpdateActuatorValueBody{
			Id:  component_id,
			Val: new_val,
		},
	)
	if err != nil {
		fmt.Printf("JSON Marshalling Error: %v\n", err)
	}

	_, err = http.Post(
		SERVER_URL+"/update_actuator_value",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		fmt.Printf("Post Request Error %v\n", err)
	}
}

type RenameSensorBody struct {
	Id   uint64 `json:"id"`
	Name string `json:"name"`
	Room string `json:"room"`
}

func RenameSensor(component_id uint64, new_name string, new_room_name string) {
	body, err := json.Marshal(
		RenameSensorBody{
			Id:   component_id,
			Name: new_name,
			Room: new_room_name,
		},
	)
	if err != nil {
		fmt.Printf("JSON Marshalling Error: %v\n", err)
	}

	_, err = http.Post(
		SERVER_URL+"/rename_component",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		fmt.Printf("Post Request Error: %v\n", err)
	}
}
