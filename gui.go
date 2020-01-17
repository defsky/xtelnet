package main

import (
	"fmt"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

var screen = tview.NewTextView().
	SetDynamicColors(true).SetScrollable(false).
	SetChangedFunc(func() {
		app.Draw()
	})

var statusBar = tview.NewTable().
	SetCell(0, 0, tview.NewTableCell(" - ").
		SetMaxWidth(40).
		SetTextColor(tcell.ColorDarkRed)).
	SetCell(0, 1, tview.NewTableCell(" - ").
		SetMaxWidth(40).
		SetTextColor(tcell.ColorDarkMagenta))

var inputBox = tview.NewInputField().SetLabel("Telnet> ").
	SetLabelColor(tcell.ColorYellow).
	SetFieldBackgroundColor(tcell.ColorDefault)

var layout = tview.NewFlex().SetDirection(tview.FlexRow).
	AddItem(screen, 0, 1, false).
	AddItem(statusBar, 1, 1, false).
	AddItem(inputBox, 1, 1, true)

// var layout = tview.NewGrid().SetRows(0, 1, 1).SetColumns(0).
// 	AddItem(screen, 0, 0, 1, 1, 1, 30, false).
// 	AddItem(statusBar, 1, 0, 1, 1, 1, 30, false).
// 	AddItem(inputBox, 2, 0, 1, 1, 1, 30, true)

func init() {
	screen.SetText("[green::b]Welcome to xtelnet!\n\n[yellow::b]Type /<Enter> for help\n\n")

	statusBar.SetBackgroundColor(tcell.ColorDarkGray)

	inputBox.SetBackgroundColor(tcell.ColorDefault)
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
				UserShell.SetSession(nil)
			} else {
				fmt.Fprintln(screen, "No session, Use /open <host> <port> to open one")
			}
		case tcell.KeyCtrlC:
			if sess := UserShell.GetSession(); sess != nil {
				sess.Close()
			}
		}
		return e
	})

}
