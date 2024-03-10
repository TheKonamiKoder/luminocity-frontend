package main

import (
	"image/color"
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

	components := []Component{}

	GetComponents(&components)
	house := make(House)
	PopulateHouseWithComponents(&house, components)

	current_room := house.GetRooms()[0]

	// ************************ DISPLAYS ************************

	SENSOR_LIST_ITEM_HEIGHT := 100
	component_listbox := widget.NewList(
		// returns total items in the list
		func() int {
			return len(house[current_room])
		}, // length

		// Creates template objects for all the items in the list, which are then
		// changed in relation to type of item
		func() fyne.CanvasObject {
			template_co := canvas.NewRectangle(color.Black)
			template_co.SetMinSize(fyne.NewSize(1, SENSOR_LIST_ITEM_HEIGHT))

			return container.NewMax(template_co)
		}, // createItem

		// *This is defined later on using a closure that has access to the rename_component_form*
		// *It had to be after the rename_component_form to be declared)*
		func(lii widget.ListItemID, co fyne.CanvasObject) {}, // updateItem
	)

	room_listbox := widget.NewList(
		func() int {
			return len(house.GetRooms())
		}, // length
		func() fyne.CanvasObject {
			return widget.NewButton("Template", func() {})
		}, // createItem
		func(lii widget.ListItemID, co fyne.CanvasObject) {
			b := co.(*widget.Button)

			room_name := house.GetRooms()[lii]

			b.SetText(room_name)
			b.OnTapped = func() {
				current_room = room_name
				component_listbox.Refresh()
			}
		}, // updateItem
	)

	left_bar := container.NewBorder(
		container.NewMax(
			container.NewBorder(
				nil,
				widget.NewSeparator(),
				widget.NewLabel("Rooms"),
				nil,
			),
		),
		////add_rdoom_form,
		nil, nil, nil,
		room_listbox,
	)

	component_name_entry := widget.NewEntry()
	component_name_entry.SetPlaceHolder("Sensor Name")
	room_select_entry := widget.NewSelectEntry(house.GetRooms())
	room_select_entry.SetPlaceHolder("Room Name")

	// The OnSubmit function will be changed by the menu with a closure that has access to the id
	// Probably not the best way to do this, but this is the simplest way that I can think of
	rename_component_form := &widget.Form{
		Items: []*widget.FormItem{
			{
				Text:   "Sensor Name:",
				Widget: component_name_entry,
			},
			{
				Text:   "Room:",
				Widget: room_select_entry,
			},
		},
	}

	rename_component_form.Hide()

	rename_component_form.OnCancel = func() { rename_component_form.Hide() }

	component_listbox.UpdateItem = func(lii widget.ListItemID, co fyne.CanvasObject) {
		c := co.(*fyne.Container)
		c.Objects[0] = newFinalSensorCanvasObject(
			house[current_room][lii],
			newSensorCavasObject(house[current_room][lii]),
			// The menu is there so that items can be renamed and delet4ed and anything else
			fyne.Menu{
				Label: "",
				Items: []*fyne.MenuItem{
					{
						Label: "Rename",
						Action: func() {
							component_name_entry.SetText(house[current_room][lii].Name)
							room_select_entry.Entry.SetText(house[current_room][lii].Room)

							rename_component_form.OnSubmit = func() {
								RenameSensor(
									house[current_room][lii].Id,
									component_name_entry.Text,
									room_select_entry.Entry.Text,
								)

								// Remove the component from that room.
								house[current_room] = append(
									house[current_room][:lii],
									house[current_room][lii+1:]...,
								)

								room_listbox.Refresh()
								component_listbox.Refresh()

								rename_component_form.Hide()
							}

							rename_component_form.Show()
						},
					},
					{
						// TODO: Make this actually work - at the moment, there is no DELETE method
						Label: "Delete",
						Action: func() {
							house[current_room] = append(
								house[current_room][:lii],
								house[current_room][lii+1:]...,
							)
						},
					},
				},
			},
		)
	}

	component_display := container.NewBorder(
		container.NewMax(
			container.NewBorder(
				nil,
				widget.NewSeparator(),
				widget.NewLabel("Sensors"),
				nil,
			),
		),
		rename_component_form,
		nil, nil,
		component_listbox,
	)

	// The component display tab
	main_content := container.NewMax(
		container.NewBorder(
			nil, nil,
			left_bar,
			nil,
			component_display,
		),
	)

	// This is the main window content
	w.SetContent(
		container.NewBorder(
			container.NewCenter(title_label),
			nil, nil, nil,
			main_content,
		),
	)

	// Asynchronous function that is responsible for keeping the components list up to date
	go func() {
		for range time.Tick(time.Second) {
			GetComponents(&components)
			PopulateHouseWithComponents(&house, components)

			// Delete empty rooms
			for room := range house {
				if len(house[room]) == 0 {
					delete(house, room)
				}
			}

			// Update the selection of rooms with all rooms
			room_select_entry.SetOptions(house.GetRooms())

			room_listbox.Refresh()
			component_listbox.Refresh()
		}
	}()

	w.ShowAndRun()
}

// ************************ COMPONENT CANVAS OBJECTS ************************

func newSensorCavasObject(component Component) fyne.CanvasObject {
	switch component.Type {
	case LED_ACTUATOR:
		return newLEDActuatorCanvasObject(component)
	case LIGHT_SENSOR:
		return newLightSensorCavasObject(component)
	case DHT11_SENSOR:
		return newDHT11SensorCanvasObject(component)
	case MOTION_SENSOR:
		return newMotionSensorCanvasObject(component)
	default:
		return widget.NewLabel("Unknown Sensor Type")
	}
}

func newLEDActuatorCanvasObject(component Component) fyne.CanvasObject {
	val := component.Data.GetVal().(bool)

	color_indication := canvas.NewRectangle(
		// The color changes from black (if it is turned off) to yellow (if the LED is turned on)
		func() color.Color {
			// Light is turned on, so display yellow
			if val {
				return color.NRGBA{
					R: 255,
					G: 255,
					B: 0,
					A: 255,
				}
			} else {
				return color.Black
			}
		}(),
	)

	text_inidication := func() *canvas.Text {
		if val {
			return canvas.NewText(
				"LED IS ON",
				color.Black,
			)
		} else {
			return canvas.NewText(
				"LED IS OFF",
				color.White,
			)
		}
	}()

	button := func() *widget.Button {
		if val {
			return widget.NewButton(
				"TURN LED OFF", // label
				func() { UpdateActuatorValue(component.Id, false) }, // tapped
			)
		} else {
			return widget.NewButton(
				"TURN LED ON", // label
				func() { UpdateActuatorValue(component.Id, true) }, //tapped
			)
		}
	}()

	led_co := container.NewGridWithRows(
		2, // rows
		container.NewMax(
			color_indication,
			container.NewCenter(text_inidication),
		),
		container.NewMax(
			button,
		),
	)

	return led_co
}

func newLightSensorCavasObject(component Component) fyne.CanvasObject {

	// Creates a rectangle with a yellow colour to display the light
	color_indication := canvas.NewRectangle(
		color.NRGBA{
			R: component.Data.GetVal().(uint8),
			G: component.Data.GetVal().(uint8),
			B: 0,
			A: 255,
		},
	)

	color_indication.SetMinSize(fyne.NewSize(1, 50))

	component_value := canvas.NewText(
		strconv.Itoa(int(component.Data.GetVal().(uint8))),
		color.White,
	)

	light_co := container.NewMax(
		color_indication,
		container.NewCenter(component_value),
	)

	return light_co
}

func newDHT11SensorCanvasObject(component Component) fyne.CanvasObject {

	temperature := component.Data.GetVal().(DHT11SensorDataVal).Temperature

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

	humidity := component.Data.GetVal().(DHT11SensorDataVal).Humidity

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

	return dht11_co
}

func newMotionSensorCanvasObject(component Component) fyne.CanvasObject {

	movement := component.Data.GetVal().(bool)

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

	return motion_co

}

type ContextMenuButton struct {
	widget.Button
	menu *fyne.Menu
}

func (b *ContextMenuButton) Tapped(e *fyne.PointEvent) {
	widget.ShowPopUpMenuAtPosition(b.menu, fyne.CurrentApp().Driver().CanvasForObject(b), e.AbsolutePosition)
}

func newFinalSensorCanvasObject(component Component, component_co fyne.CanvasObject, menu fyne.Menu) fyne.CanvasObject {

	component_type_name := component.Type.GetName()

	final_co := container.NewMax(
		container.NewBorder(
			container.NewMax(
				container.NewBorder(
					nil, nil,
					widget.NewLabel(component.Name+" ("+component_type_name+")"),
					&ContextMenuButton{
						Button: *widget.NewButtonWithIcon(
							"", theme.MenuIcon(), func() {}),
						menu: &menu,
					},
				),
			),
			nil, nil, nil,
			component_co,
		),
	)

	return final_co
}
