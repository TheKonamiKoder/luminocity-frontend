package main

type SensorType int

const (
	LIGHT_SENSOR = iota
	TEMPERATURE_SENSOR
	HUMIDITY_SENSOR
	MOTION_SENSOR
)

// LightSensorData | TemperatureSensorData | HumiditySensorData | MotionSensorData
type SensorData interface{ GetVal() interface{} }

type LightSensorData struct{ Val uint8 }

func (l LightSensorData) GetVal() interface{} {
	return l.Val
}

type TemperatureSensorData struct{ Val float32 }

func (t TemperatureSensorData) GetVal() interface{} {
	return t.Val
}

type HumiditySensorData struct{ Val float32 }

func (h HumiditySensorData) GetVal() interface{} {
	return h.Val
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
