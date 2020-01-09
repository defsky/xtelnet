package main

import (
	"fmt"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()
	screen := tview.NewTextView().SetDynamicColors(true)

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
			msg, data, err := doCommand(inputBox.GetText())
			inputBox.SetText("")
			if err != nil {
				app.QueueUpdateDraw(func() {
					fmt.Fprintf(screen, "[red]%s\n", err)
				})
			}
			if len(msg) > 0 {
				app.QueueUpdateDraw(func() {
					fmt.Fprintf(screen, "%s\n", msg)
				})
			}
			if len(data) > 0 {
				app.QueueUpdateDraw(func() {
					fmt.Fprintln(screen, fmt.Sprintf("send data: %v", data))
				})
			}
		case tcell.KeyEsc:
			inputBox.SetText("")
		case tcell.KeyUp:
		case tcell.KeyDown:
		case tcell.KeyTab, tcell.KeyBacktab:
		}
	})

	layout := tview.NewGrid().SetRows(0, 1, 1).SetColumns(0).
		AddItem(screen, 0, 0, 1, 1, 1, 30, false).
		AddItem(statusBar, 1, 0, 1, 1, 1, 30, false).
		AddItem(inputBox, 2, 0, 1, 1, 1, 30, true)

	app.SetRoot(layout, true)
	if err := app.Run(); err != nil {
		panic(err)
	}
}
