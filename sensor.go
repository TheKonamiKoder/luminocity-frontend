package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
)

const SERVER_URL = "***************"

type SensorType int

const (
	LIGHT_SENSOR = iota
	DHT11_SENSOR
	MOTION_SENSOR
)

func (st SensorType) GetName() string {
	switch st {
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

// LightSensorData | TemperatureSensorData | HumiditySensorData | MotionSensorData
type SensorData interface {
	GetVal() interface{}
	SetVal(Val interface{})
}

type Sensor struct {
	Id   uint64
	Name string
	Room string
	Type SensorType
	Data SensorData
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

type House map[string][]Sensor

func PopulateHouseWithSensors(house *House, sensors []Sensor) {
	var sensor Sensor
	for i := 0; i < len(sensors); i++ {
		sensor = sensors[i]

		sensors_in_room, room_exists := (*house)[sensor.Room]

		if !room_exists {
			(*house)[sensor.Room] = []Sensor{sensor}
			continue
		}

		////house[sensor.Room] = append(sensors_in_room, sensor)

		sensor_is_in_room := false
		for j := 0; j < len(sensors_in_room); j++ {
			if sensors_in_room[j].Id == sensor.Id {
				sensors_in_room[j].Name = sensor.Name
				sensors_in_room[j].Data.SetVal(sensor.Data.GetVal())
				sensor_is_in_room = true
			}
		}

		if !sensor_is_in_room {
			(*house)[sensor.Room] = append((*house)[sensor.Room], sensor)
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

// The JsonSensor is used for parsing the json recieved from the server
// It is very similar to the Sensor struct and the only reason why it exists is because
// of the Data and Val attributes, as they are different for each of the sensor types
// The server also has all of the attributes in lower case and instead of changing the
// API based on the product (Go only exports fields with capital letters, but python
// doesn't care about the capital letters), I decided to use the struct tags
type JsonSensor struct {
	Id   uint64      `json:"id"`
	Name string      `json:"name"`
	Room string      `json:"room"`
	Type SensorType  `json:"type"`
	Val  interface{} `json:"val"`
}

func UpdateSensorValues(sensors *[]Sensor) {
	resp, err := http.Get(SERVER_URL + "/get_all")
	if err != nil {
		fmt.Printf("Get Request Error: %v\n", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Reading Response Error: %v\n", err)
	}

	var sensors_json []JsonSensor

	err = json.Unmarshal(body, &sensors_json)
	if err != nil {
		fmt.Printf("Json Unmarshaling Error: %v\n", err)
		fmt.Printf("body: %v\n", string(body))
	}

	for i, json_sensor := range sensors_json {
		if i == len(*sensors) {
			*sensors = append(*sensors, jsonSensorToSensor(json_sensor))
			continue
		}

		sensor := &(*sensors)[i]

		// A sensor's id will never change, even though it's name may.
		// A sensor's type will also never change.
		// If a sensor does not have same id as json_sensor's id, it must have been deleted.
		if sensor.Id != json_sensor.Id {
			// Removes the sensor by creating a new slice without it
			*sensors = append((*sensors)[:i], (*sensors)[i+1:]...)
			continue
		}

		// It is possible for a sensor's name, room and value to change
		sensor.Name = json_sensor.Name
		sensor.Room = json_sensor.Room
		sensor.Data.SetVal(jsonSensorToSensor(json_sensor).Data.GetVal())
	}
}

func jsonSensorToSensor(json_sensor JsonSensor) Sensor {
	return Sensor{
		Id:   json_sensor.Id,
		Name: json_sensor.Name,
		Room: json_sensor.Room,
		Type: json_sensor.Type,
		Data: func() SensorData {
			switch json_sensor.Type {
			case LIGHT_SENSOR:
				return &LightSensorData{
					// The json library casts every numeric item to a float64 apparently
					Val: uint8(json_sensor.Val.(float64)),
				}
			case DHT11_SENSOR:
				return &DHT11SensorData{
					Val: DHT11SensorDataVal{
						// The json library also casts arrays to []interface{} also...
						Temperature: float32(json_sensor.Val.([]interface{})[0].(float64)),
						Humidity:    float32(json_sensor.Val.([]interface{})[1].(float64)),
					},
				}
			case MOTION_SENSOR:
				return &MotionSensorData{
					Val: json_sensor.Val.(bool),
				}
			default:
				panic("Unknown sensor type")
			}
		}(),
	}
}

type RenameSensorBody struct {
	Id   uint64 `json:"id"`
	Name string `json:"name"`
	Room string `json:"room"`
}

func RenameSensor(sensor_id uint64, new_name string, new_room_name string) {
	body, err := json.Marshal(
		RenameSensorBody{
			Id:   sensor_id,
			Name: new_name,
			Room: new_room_name,
		},
	)
	if err != nil {
		fmt.Printf("JSON Marshalling Error: %v\n", err)
	}

	_, err = http.Post(
		SERVER_URL+"/rename_sensor",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		fmt.Printf("Post Request Error: %v\n", err)
	}
}
