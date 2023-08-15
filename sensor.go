package main

type SensorType int

const (
	LIGHT_SENSOR = iota
	DHT11_SENSOR
	MOTION_SENSOR
)

// LightSensorData | TemperatureSensorData | HumiditySensorData | MotionSensorData
type SensorData interface{ GetVal() interface{} }

type LightSensorData struct{ Val uint8 }

func (l LightSensorData) GetVal() interface{} {
	return l.Val
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

type MotionSensorData struct{ Val bool }

func (m MotionSensorData) GetVal() interface{} {
	return m.Val
}

type Sensor struct {
	Id   uint32
	Name string
	Room string
	Type SensorType
	Data SensorData
}
