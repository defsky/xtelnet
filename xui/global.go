package xui

import (
	"sort"
	"strings"
	"xtelnet/session"

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

const historyCmdLength = 1000

var historyCmd = session.NewHistoryCmd(historyCmdLength)
var inputCh = make(chan []byte, 10)

func init() {
	historyCmd.LoadCache()

	// screen.SetText("[green]Welcome to xtelnet!\n\n[yellow]Type /<Enter> for help\n\n[-]")
	screen.SetDrawFunc(func(scr tcell.Screen, x int, y int, width int, height int) (int, int, int, int) {
		if width < 110 {
			screen.SetWrap(false)
		} else {
			screen.SetWrap(true)
		}
		return x, y, width, height
	})

	statusBar.SetBackgroundColor(tcell.ColorDarkGray)

	inputBox.SetBackgroundColor(tcell.ColorDefault)
	inputBox.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			cmdstr := inputBox.GetText()
			inputBox.SetText("")
			historyCmd.Add(cmdstr)

			inputCh <- []byte(cmdstr)
		case tcell.KeyEsc:
			inputBox.SetText("")
		case tcell.KeyTab, tcell.KeyBacktab:
		}
	})

	words := []string{}
	for _, c := range session.GetRootCmd().GetCommandMap() {
		words = append(words, c.Name()+" ")
	}
	sort.Strings(words)
	inputBox.SetAutocompleteFunc(func(currentText string) (entries []string) {
		if len(currentText) == 0 {
			return
		}
		for _, word := range words {
			if strings.HasPrefix(strings.ToLower(word), strings.ToLower(currentText)) {
				entries = append(entries, word)
			}
		}
		if entries != nil {
			entries = append(entries, "")
		}
		return
	})

	inputBox.SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
		key := e.Key()

		switch key {
		case tcell.KeyUp, tcell.KeyDown:
			if !historyCmd.IsScrolling() {
				historyCmd.SetScrolling(true)
				historyCmd.SetCurrentText(inputBox.GetText())
				historyCmd.Match()
			}
			s := historyCmd.NextMatch(key)
			if len(s) > 0 {
				s = strings.Trim(s, " ")
				inputBox.SetText(s)
			}
		default:
			if historyCmd.IsScrolling() {
				historyCmd.SetScrolling(false)
			}
		}

		return e
	})
}
