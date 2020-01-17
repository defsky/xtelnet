package main

import "github.com/rivo/tview"

// gui app
var app = tview.NewApplication()

func main() {
	app.SetRoot(layout, true)
	if err := app.Run(); err != nil {
		panic(err)
	}
}
