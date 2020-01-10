package main

import (
	"fmt"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

// gui app
var app = tview.NewApplication()

var screen = tview.NewTextView().
	SetDynamicColors(true).
	SetChangedFunc(func() {
		app.Draw()
	})

var UserShell = &Shell{}

func main() {

	hostCell := tview.NewTableCell("No active connection").
		SetMaxWidth(40).
		SetTextColor(tcell.ColorDarkRed)
	statusCell := tview.NewTableCell(" - ").
		SetMaxWidth(40).
		SetTextColor(tcell.ColorDarkMagenta)
	statusBar := tview.NewTable().
		SetCell(0, 0, hostCell).
		SetCell(0, 1, statusCell)
	statusBar.SetBackgroundColor(tcell.ColorDarkGray)

	inputBox := tview.NewInputField().SetLabel("Telnet> ").
		SetFieldBackgroundColor(tcell.ColorDefault)
	inputBox.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			msg, err := UserShell.Exec(inputBox.GetText())
			inputBox.SetText("")
			if err != nil {
				app.QueueUpdate(func() {
					fmt.Fprintf(screen, "[red]%s\n", err)
				})
			}
			if len(msg) > 0 {
				app.QueueUpdate(func() {
					fmt.Fprintf(screen, "%s\n", msg)
				})
			}
		case tcell.KeyEsc:
			inputBox.SetText("")
		case tcell.KeyUp:
		case tcell.KeyDown:
		case tcell.KeyTab, tcell.KeyBacktab:
		}
	})

	app.SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
		key := e.Key()
		switch key {
		case tcell.KeyCtrlD:
			if sess := UserShell.GetSession(); sess != nil {
				sess.Close()
			} else {
				fmt.Fprintln(screen, "No session, Use /open <host> <port> to open one")
			}
		}
		return e
	})
	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(screen, 0, 1, false).
		AddItem(statusBar, 1, 1, false).
		AddItem(inputBox, 1, 1, true)
	// layout := tview.NewGrid().SetRows(0, 1, 1).SetColumns(0).
	// 	AddItem(screen, 0, 0, 1, 1, 1, 30, false).
	// 	AddItem(statusBar, 1, 0, 1, 1, 1, 30, false).
	// 	AddItem(inputBox, 2, 0, 1, 1, 1, 30, true)

	app.SetRoot(layout, true)
	if err := app.Run(); err != nil {
		panic(err)
	}
}
