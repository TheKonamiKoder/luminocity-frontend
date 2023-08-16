package main

import (
	"image/color"
	"strconv"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/container"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("Luminocity")

	title_label := widget.NewLabel("Luminocity")

	sensors := [...]Sensor{
		{
			Id:   0,
			Name: "Ceiling Lights",
			Room: "Kitchen",
			Type: LIGHT_SENSOR,
			Data: LightSensorData{Val: 10},
		},
		{
			Id:   1,
			Name: "Lamp",
			Room: "Hall",
			Type: LIGHT_SENSOR,
			Data: LightSensorData{Val: 255},
		},
		{
			Id:   1,
			Name: "Sunlight",
			Room: "Outside",
			Type: LIGHT_SENSOR,
			Data: LightSensorData{Val: 50},
		},
		{
			Id:   3,
			Name: "Loft",
			Room: "Loft",
			Type: DHT11_SENSOR,
			Data: DHT11SensorData{Val: DHT11SensorDataVal{38, 80}},
		},
		{
			Id:   4,
			Name: "Basement",
			Room: "Basement",
			Type: DHT11_SENSOR,
			Data: DHT11SensorData{Val: DHT11SensorDataVal{0, 12}},
		},
		{
			Id:   5,
			Name: "Sun",
			Room: "Outside",
			Type: DHT11_SENSOR,
			Data: DHT11SensorData{Val: DHT11SensorDataVal{50, 90}},
		},
		{
			Id:   6,
			Name: "Front door",
			Room: "Outside",
			Type: MOTION_SENSOR,
			Data: MotionSensorData{Val: true},
		},
		{
			Id:   7,
			Name: "Back door",
			Room: "Garden",
			Type: MOTION_SENSOR,
			Data: MotionSensorData{Val: false},
		},
	}

	LIST_ITEM_HEIGHT := 100
	sensor_displays := widget.NewList(
		func() int {
			return len(sensors)
		},
		func() fyne.CanvasObject {
			template_co := canvas.NewRectangle(color.Black)

			template_co.SetMinSize(fyne.NewSize(1, LIST_ITEM_HEIGHT))

			return container.NewMax(template_co)

			//return sco
		},
		func(lii widget.ListItemID, co fyne.CanvasObject) {
			c := co.(*fyne.Container)
			c.Objects[0] = newSensorCavasObject(sensors[lii])
		},
	)

	// The sensor display tab
	sensor_display := container.NewTabItem(
		"Sensors",
		container.NewMax(sensor_displays),
	)

	// * This is the main window content
	w.SetContent(
		container.NewBorder(
			container.NewCenter(title_label),
			nil, nil, nil,
			container.NewAppTabs(sensor_display),
		),
	)

	w.ShowAndRun()
}

func newSensorCavasObject(sensor Sensor) fyne.CanvasObject {
	switch sensor.Type {
	case LIGHT_SENSOR:
		return newLightSensorCavasObject(sensor)
	case DHT11_SENSOR:
		return newDHT11SensorCanvasObject(sensor)
	case MOTION_SENSOR:
		return newMotionSensorCanvasObject(sensor)
	default:
		return widget.NewLabel("Template")
	}
}

func newLightSensorCavasObject(sensor Sensor) fyne.CanvasObject {

	// Creates a rectangle with a yellow colour to display the light
	color_indication := canvas.NewRectangle(
		color.NRGBA{
			R: sensor.Data.GetVal().(uint8),
			G: sensor.Data.GetVal().(uint8),
			B: 0,
			A: 255,
		},
	)

	color_indication.SetMinSize(fyne.NewSize(1, 50))

	sensor_value := canvas.NewText(
		strconv.Itoa(int(sensor.Data.GetVal().(uint8))),
		color.White,
	)

	light_co := container.NewMax(
		color_indication,
		container.NewCenter(sensor_value),
	)

	return newFinalSensorCanvasObject(sensor, light_co)
}

func newDHT11SensorCanvasObject(sensor Sensor) fyne.CanvasObject {

	temperature := sensor.Data.GetVal().(DHT11SensorDataVal).Temperature

	temperature_color_indication := canvas.NewRectangle(
		// The color changes from red (if it is above 25) or blue (below 25) to show hot and cold.
		func() color.Color {
			BASE := uint8(100)

			RED_CUTOFF := float32(25)
			MAX_TEMP := float32(60) // All float32 to avoid MismatchedTypes
			MIN_TEMP := float32(-20)

			if temperature >= float32(RED_CUTOFF) {
				temp_percent := (temperature - RED_CUTOFF) / (MAX_TEMP - RED_CUTOFF)
				return color.NRGBA{
					R: 255,
					G: BASE - uint8((1-temp_percent)*(MAX_TEMP-RED_CUTOFF)),
					B: BASE - uint8((1-temp_percent)*(MAX_TEMP-RED_CUTOFF)),
					A: 255,
				}
			} else {
				temp_percent := (RED_CUTOFF - temperature) / (RED_CUTOFF - MIN_TEMP)
				return color.NRGBA{
					R: BASE - uint8(temp_percent*(MAX_TEMP-RED_CUTOFF)),
					G: BASE - uint8(temp_percent*(MAX_TEMP-RED_CUTOFF)),
					B: 255,
					A: 255,
				}
			}
		}(),
	)

	temperature_text_indication := canvas.NewText(
		strconv.FormatFloat(float64(temperature), 'f', 2, 64)+"°C",
		color.White,
	)

	humidity := sensor.Data.GetVal().(DHT11SensorDataVal).Humidity

	// Just a standard thingy...
	humidity_color_indication := canvas.NewRectangle(
		color.NRGBA{
			R: 5,
			G: 238,
			B: 255,
			A: 255,
		},
	)

	humidity_text_indication := canvas.NewText(
		strconv.FormatFloat(float64(humidity), 'f', 2, 64)+"%",
		color.White,
	)

	dht11_co := container.NewGridWithColumns(
		2,
		container.NewMax(
			temperature_color_indication,
			container.NewCenter(temperature_text_indication),
		),
		container.NewMax(
			humidity_color_indication,
			container.NewCenter(humidity_text_indication),
		),
	)

	return newFinalSensorCanvasObject(sensor, dht11_co)
}

func newMotionSensorCanvasObject(sensor Sensor) fyne.CanvasObject {

	movement := sensor.Data.GetVal().(bool)

	color_indication := canvas.NewRectangle(
		// Displays green if there is no movement and red if there is movement
		func() color.Color {
			if !movement {
				return color.NRGBA{
					R: 76,
					G: 235,
					B: 52,
					A: 255,
				}
			} else {
				return color.NRGBA{
					R: 235,
					G: 76,
					B: 52,
					A: 255,
				}
			}
		}(),
	)

	text_indication := canvas.NewText(
		func() string {
			if !movement {
				return "NO MOVEMENT DETECTED"
			} else {
				return "MOVEMENT DETECTED"
			}
		}(),
		color.White,
	)

	motion_co := container.NewMax(
		color_indication,
		container.NewCenter(text_indication),
	)

	return newFinalSensorCanvasObject(sensor, motion_co)

}

// Wraps the main part of the sensor in the repetitive part
func newFinalSensorCanvasObject(sensor Sensor, sensor_co fyne.CanvasObject) fyne.CanvasObject {

	sensor_type_name := func() string {
		switch sensor.Type {
		case LIGHT_SENSOR:
			return "Light"
		case DHT11_SENSOR:
			return "DHT11"
		case MOTION_SENSOR:
			return "Motion"
		default:
			return "Unknown"
		}
	}()

	final_co := container.NewMax(
		container.NewBorder(
			container.NewMax(
				container.NewBorder(
					nil, nil,
					widget.NewLabel(sensor.Name+" ("+sensor_type_name+")"),
					widget.NewButtonWithIcon(
						"",
						theme.MenuIcon(),
						func() {}, // TODO: Show option to see log and analytics
					),
				),
			),
			nil, nil, nil,
			sensor_co,
		),
	)

	return final_co
}
