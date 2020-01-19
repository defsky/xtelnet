package main

import (
	"container/list"
	"fmt"
	"sort"
	"strings"

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

const historyCmdListLength = 1000

type HistoryCmdList struct {
	history          *list.List
	matchs           *list.List
	length           int
	currentInputText string
	lastInputText    string
	currentMatch     *list.Element
}

func NewHistoryCmdList(len int) *HistoryCmdList {

	return &HistoryCmdList{
		length:  historyCmdListLength,
		history: list.New(),
		matchs:  list.New(),
	}
}
func (l *HistoryCmdList) SetCurrentText(text string) {
	l.currentInputText = text
}
func (l *HistoryCmdList) NextMatch() string {
	if l.currentInputText != l.lastInputText {
		l.match(l.currentInputText)
		l.currentMatch = l.matchs.Front()
	}
	if len(l.currentInputText) == 0 {
		l.matchs = l.history
		l.currentMatch = l.matchs.Front()
	}
	if l.matchs.Len() > 0 {
		s := l.currentMatch.Value.(string)
		l.currentMatch = l.currentMatch.Next()
		return s
	}
	return ""
}
func (l *HistoryCmdList) PrevMatch() string {
	if l.currentInputText != l.lastInputText {
		l.match(l.currentInputText)
	}
	if len(l.currentInputText) == 0 {
		l.matchs = l.history
		l.currentMatch = l.matchs.Front()
	}
	if l.matchs.Len() > 0 {
		s := l.currentMatch.Value.(string)
		l.currentMatch = l.currentMatch.Prev()
		return s
	}
	return ""
}
func (l *HistoryCmdList) match(text string) {
	l.matchs.Init()
	l.currentMatch = nil
	l.currentInputText = text
	for e := l.history.Front(); e != nil; e.Next() {
		s := e.Value.(string)
		if strings.HasPrefix(strings.ToLower(s), strings.ToLower(text)) {
			l.matchs.PushFront(e)
		}
	}
	if l.matchs.Len() > 0 {
		l.currentMatch = l.matchs.Front()
	}
}

// Add text into history record
func (l *HistoryCmdList) Add(text string) {
	// if exists, move to front
	for e := l.history.Front(); e != nil; e.Next() {
		s := e.Value.(string)
		if s == text {
			l.history.MoveToFront(e)
		}
		return
	}

	// add new record
	if l.history.Len() > l.length {
		lastElement := l.history.Back()
		l.history.Remove(lastElement)
	}
	l.history.PushFront(text)
}

var historyCmdList = NewHistoryCmdList(historyCmdListLength)

func init() {
	screen.SetText("[green::b]Welcome to xtelnet!\n\n[yellow::b]Type /<Enter> for help\n\n")
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
			historyCmdList.Add(cmdstr)
			msg, err := UserShell.Exec(cmdstr)
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

	words := []string{}
	for _, c := range rootCMD.subCommand {
		words = append(words, c.name+" ")
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
	inputBox.SetChangedFunc(updateHist)
	inputBox.SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
		switch e.Key() {
		case tcell.KeyUp:
			s := historyCmdList.NextMatch()
			if len(s) > 0 {
				inputBox.SetChangedFunc(nil)
				inputBox.SetText(s)
				inputBox.SetChangedFunc(updateHist)
			}
		case tcell.KeyDown:
			s := historyCmdList.PrevMatch()
			if len(s) > 0 {
				inputBox.SetChangedFunc(nil)
				inputBox.SetText(s)
				inputBox.SetChangedFunc(updateHist)
			}
		}
		return e
	})
}

func updateHist(text string) {
	historyCmdList.SetCurrentText(text)
}
