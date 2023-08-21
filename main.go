package main

import (
	"image/color"
	"math/rand"
	"sort"
	"strconv"
	"time"

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

	sensors := []Sensor{}
	house := make(map[string][]Sensor)

	rooms := make([]string, 0, len(house))
	////rooms := make([]string, 0, len(house))

	UpdateSensorValues(&sensors)

	var sensor Sensor
	for i := 0; i < len(sensors); i++ {
		sensor = sensors[i]

		sensors_in_room, room_exists := house[sensor.Room]

		if !room_exists {
			house[sensor.Room] = []Sensor{sensor}
			continue
		}

		house[sensor.Room] = append(sensors_in_room, sensor)
	}

	for r := range house {
		rooms = append(rooms, r)
	}
	sort.Strings(rooms)

	current_room := rooms[0]
	current_sensors := house[current_room]

	SENSOR_LIST_ITEM_HEIGHT := 100
	sensor_listbox := widget.NewList(
		func() int {
			return len(current_sensors)
		},
		func() fyne.CanvasObject {
			template_co := canvas.NewRectangle(color.Black)
			template_co.SetMinSize(fyne.NewSize(1, SENSOR_LIST_ITEM_HEIGHT))

			return container.NewMax(template_co)
		},
		func(lii widget.ListItemID, co fyne.CanvasObject) {
			c := co.(*fyne.Container)
			c.Objects[0] = newSensorCavasObject(current_sensors[lii])
		},
	)

	room_listbox := widget.NewList(
		func() int {
			return len(rooms)
		},
		func() fyne.CanvasObject {
			return widget.NewButton("Template", func() {})
		},
		func(lii widget.ListItemID, co fyne.CanvasObject) {
			b := co.(*widget.Button)

			room_name := rooms[lii]

			b.SetText(room_name)
			b.OnTapped = func() {
				current_room = room_name
				current_sensors = house[current_room]
				sensor_listbox.Refresh()
			}
		},
	)

	room_name_entry := widget.NewEntry()

	add_room_form := &widget.Form{
		Items: []*widget.FormItem{
			{
				Text:   "Name: ",
				Widget: room_name_entry,
			},
		},
	}
	add_room_form.Hide()

	add_room_form.OnSubmit = func() {
		new_room_name := room_name_entry.Text

		house[new_room_name] = []Sensor{}
		rooms = append(rooms, new_room_name)
		sort.Strings(rooms)
		room_listbox.Refresh()

		add_room_form.Hide()
	}

	add_room_form.OnCancel = func() { add_room_form.Hide() }

	left_bar := container.NewBorder(
		container.NewMax(
			container.NewBorder(
				nil,
				widget.NewSeparator(),
				widget.NewLabel("Rooms"),
				widget.NewButtonWithIcon(
					"",
					theme.ContentAddIcon(),
					func() {
						add_room_form.Show()
					},
				),
			),
		),
		add_room_form,
		nil, nil,
		room_listbox,
	)

	sensor_name_entry := widget.NewEntry()
	sensor_type_select := widget.NewSelect(
		[]string{"Light Sensor", "DHT11 Sensor", "Motion Sensor"},
		func(s string) {},
	)

	add_sensor_form := &widget.Form{
		Items: []*widget.FormItem{
			{
				Text:   "Sensor Name:",
				Widget: sensor_name_entry,
			},
			{
				Text:   "Sensor Type:",
				Widget: sensor_type_select,
			},
		},
	}

	add_sensor_form.OnSubmit = func() {
		new_sensor_name := sensor_name_entry.Text

		var (
			new_sensor_type SensorType
			new_sensor_data SensorData
		)

		switch sensor_type_select.Selected {
		case "Light Sensor":
			new_sensor_type = LIGHT_SENSOR
			new_sensor_data = &LightSensorData{Val: uint8(rand.Intn(255))} // Data is random for now
		case "DHT11 Sensor":
			new_sensor_type = DHT11_SENSOR
			new_sensor_data = &DHT11SensorData{Val: DHT11SensorDataVal{
				Temperature: (rand.Float32() * 80) - 20,
				Humidity:    rand.Float32() * 100,
			}}
		case "Motion Sensor":
			new_sensor_type = MOTION_SENSOR
			new_sensor_data = &MotionSensorData{Val: rand.Intn(2) != 0}
		}

		house[current_room] = append(house[current_room], Sensor{
			Id:   100,
			Name: new_sensor_name,
			Room: current_room,
			Type: new_sensor_type,
			Data: new_sensor_data,
		})
		current_sensors = house[current_room]

		sensor_listbox.Refresh()

		add_sensor_form.Hide()
	}
	add_sensor_form.Hide()

	add_sensor_form.OnCancel = func() { add_sensor_form.Hide() }

	sensor_display := container.NewBorder(
		container.NewMax(
			container.NewBorder(
				nil,
				widget.NewSeparator(),
				widget.NewLabel("Sensors"),
				widget.NewButtonWithIcon(
					"",
					theme.ContentAddIcon(),
					func() {
						add_sensor_form.Show()
					},
				),
			),
		),
		add_sensor_form,
		nil, nil,
		sensor_listbox,
	)

	// The sensor display tab
	main_content := container.NewMax(
		container.NewBorder(
			nil, nil,
			left_bar,
			nil,
			sensor_display,
		),
	)

	//* This is the main window content
	w.SetContent(
		container.NewBorder(
			container.NewCenter(title_label),
			nil, nil, nil,
			main_content,
		),
	)

	go func() {
		for range time.Tick(time.Millisecond) {
			UpdateSensorValues(&sensors)

			var sensor Sensor
			for i := 0; i < len(sensors); i++ {
				sensor = sensors[i]

				sensors_in_room, room_exists := house[sensor.Room]

				if !room_exists {
					house[sensor.Room] = []Sensor{sensor}
					continue
				}

				////house[sensor.Room] = append(sensors_in_room, sensor)

				t := false
				for j := 0; j < len(sensors_in_room); j++ {
					if sensors_in_room[j].Id == sensor.Id {
						sensors_in_room[j].Name = sensor.Name
						sensors_in_room[j].Data.SetVal(sensor.Data.GetVal())
						t = true
					}
				}

				if !t {
					house[sensor.Room] = append(house[sensor.Room], sensor)
				}

			}

			for r := range house {
				// If the room is not found in the list of rooms, it will be appended
				if sort.SearchStrings(rooms, r) == len(rooms) {
					rooms = append(rooms, r)
				}

			}

			sort.Strings(rooms)

			room_listbox.Refresh()
			sensor_listbox.Refresh()
		}
	}()

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
