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
	}

	LIST_ITEM_HEIGHT := 100
	sensor_displays := widget.NewList(
		func() int {
			return len(sensors)
		},
		func() fyne.CanvasObject {
			r := canvas.NewRectangle(color.Black)

			r.SetMinSize(fyne.NewSize(1, LIST_ITEM_HEIGHT))

			return container.NewMax(r)

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

	// The text to show the actual value of the sensor, so it is easy to read
	sensor_value := canvas.NewText(
		strconv.Itoa(int(sensor.Data.GetVal().(uint8))),
		color.White,
	)

	visual_light_sensor := container.NewMax(
		container.NewBorder(
			container.NewMax(
				container.NewBorder(
					nil, nil,
					widget.NewLabel(sensor.Name+" (Light Sensor)"),
					widget.NewButtonWithIcon(
						"",
						theme.MenuIcon(),
						func() {}, // TODO: Show option to see log and analytics
					),
				),
			),
			nil, nil, nil,
			color_indication,
			container.NewCenter(sensor_value),
		),
	)

	return visual_light_sensor
}
