package main

type SensorType int

const (
	LIGHT_SENSOR = iota
	TEMPERATURE_SENSOR
	HUMIDITY_SENSOR
	MOTION_SENSOR
)

// LightSensorData | TemperatureSensorData | HumiditySensorData | MotionSensorData
type SensorData interface{}

type LightSensorData struct{ Val uint8 }
type TemperatureSensorData struct{ Val float32 }
type HumiditySensorData struct{ Val float32 }
type MotionSensorData struct{ Val bool }

type Sensor struct {
	Id   uint32
	Name string
	Room string
	Type SensorType
	Data SensorData
}
