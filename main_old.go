package main

import (
	"fmt"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

// gui app
var app = tview.NewApplication()

var widgets = []tview.Primitive{screen, inputBox}
var currentWidget = 1

func main_old() {
	defer historyCmd.Cache()

	app.SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
		key := e.Key()
		switch key {
		case tcell.KeyCtrlD:
			if sess := UserShell.GetSession(); sess != nil {
				sess.Close()
				UserShell.SetSession(nil)
			} else {
				fmt.Fprintln(screen, "No session, Use /open <host> <port> to open one")
			}
		case tcell.KeyCtrlC:
			if sess := UserShell.GetSession(); sess != nil {
				sess.Close()
			}

		case tcell.KeyTab:
			currentWidget = (currentWidget + 1) % len(widgets)
			p := widgets[currentWidget]
			app.SetFocus(p)
		}
		return e
	})

	app.SetRoot(layout, true)
	if err := app.Run(); err != nil {
		panic(err)
	}
}
