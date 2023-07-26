package main

import (
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/container"
	"fyne.io/fyne/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("Luminocity")

	title_label := widget.NewLabel("Luminocity")

	sensors := [...]string{"Kitchen", "Hall", "Garden"}
	sensor_displays := widget.NewList(
		func() int {
			return len(sensors)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("New Sensor")
		},
		func(lii widget.ListItemID, co fyne.CanvasObject) {
			co.(*widget.Label).SetText(sensors[lii])
		},
	)

	sensor_display := container.NewTabItem(
		"Sensors",
		container.NewBorder(
			title_label, nil, nil, nil,
			sensor_displays,
		),
	)

	w.SetContent(container.NewAppTabs(sensor_display))

	w.ShowAndRun()
}
