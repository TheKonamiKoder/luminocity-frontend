package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type SensorType int

const SERVER_URL = "http://localhost:5000"

const (
	LIGHT_SENSOR = iota
	DHT11_SENSOR
	MOTION_SENSOR
)

// LightSensorData | TemperatureSensorData | HumiditySensorData | MotionSensorData
type SensorData interface {
	GetVal() interface{}
	SetVal(Val interface{})
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

type Sensor struct {
	Id   uint32
	Name string
	Room string
	Type SensorType
	Data SensorData
}

// The JsonSensor is used for parsing the json recieved from the server
// It is very similar to the Sensor struct and the only reason why it exists is because
// of the Data and Val attributes, as they are different for each of the sensor types
// The server also has all of the attributes in lower case and instead of changing the
// API based on the product (Go only exports fields with capital letters, but python
// doesn't care about the capital letters), I decided to use the struct tags
type JsonSensor struct {
	Id   uint32     `json:"id"`
	Name string     `json:"name"`
	Room string     `json:"room"`
	Type SensorType `json:"type"`
	Data struct {
		Val interface{} `json:"val"`
	} `json:"data"`
}

func UpdateSensorValues(sensors *[]Sensor) {
	resp, err1 := http.Get("http://localhost:5000/get_all")
	if err1 != nil {
		fmt.Printf("err1: %v\n", err1)
	}

	defer resp.Body.Close()

	body, err2 := io.ReadAll(resp.Body)
	if err2 != nil {
		fmt.Printf("err2: %v\n", err2)
	}

	var sensors_json []JsonSensor

	err3 := json.Unmarshal(body, &sensors_json)
	if err3 != nil {
		fmt.Printf("err3: %v\n", err3)
	}

	//// fmt.Printf("sensors_json: %v\n", sensors_json)
	//// if len(sensors_json) == 0 {
	//// 	fmt.Printf("sensors_json is 0")
	//// }

	for i, json_sensor := range sensors_json {
		fmt.Printf("i: %v\n", i)

		if i == len(*sensors) {
			fmt.Printf("jsonSensorToSensor(json_sensor): %v\n", jsonSensorToSensor(json_sensor))
			*sensors = append(*sensors, jsonSensorToSensor(json_sensor))
			continue
		}

		fmt.Printf("(*sensors): %v\n", (*sensors))

		sensor := &(*sensors)[i]

		// A sensor's id will never change, even though it's name may.
		// A sensor's room and type will also never change.
		// If a sensor does not have same id as json_sensor's id, it must have been deleted.
		if sensor.Id != json_sensor.Id {
			// Removes the sensor by creating a new slice without it
			*sensors = append((*sensors)[:i], (*sensors)[i+1:]...)
			continue
		}

		// It is possible for a sensor's name and value to change
		sensor.Name = json_sensor.Name
		sensor.Data.SetVal(json_sensor.Data.Val)
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
					Val: uint8(json_sensor.Data.Val.(float64)),
				}
			case DHT11_SENSOR:
				return &DHT11SensorData{
					Val: DHT11SensorDataVal{
						// The json library also casts arrays to []interface{} also...
						Temperature: float32(json_sensor.Data.Val.([]interface{})[0].(float64)),
						Humidity:    float32(json_sensor.Data.Val.([]interface{})[1].(float64)),
					},
				}
			case MOTION_SENSOR:
				return &MotionSensorData{
					Val: json_sensor.Data.Val.(bool),
				}
			default:
				panic("Unknown sensor type")
			}
		}(),
	}
}
