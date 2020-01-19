package main

import (
	"container/list"
	"fmt"
	"sort"
	"strings"
	"sync"

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

const historyCmdLength = 1000

type HistoryCmd struct {
	sync.Mutex
	Scrolling        bool
	history          *list.List
	matchs           *list.List
	maxLen           int
	currentInputText string
	lastInputText    string
	currentMatch     *list.Element
}

func NewHistoryCmd(len int) *HistoryCmd {

	return &HistoryCmd{
		maxLen:  historyCmdLength,
		history: list.New(),
		matchs:  list.New(),
	}
}
func (l *HistoryCmd) SetCurrentText(text string) {
	l.Lock()
	defer l.Unlock()
	l.currentInputText = text
}

func (l *HistoryCmd) NextMatch(key tcell.Key) string {
	l.Lock()
	defer l.Unlock()

	l.match()
	if l.currentMatch != nil {
		s := l.currentMatch.Value.(string)
		switch key {
		case tcell.KeyUp:
			if n := l.currentMatch.Next(); n != nil {
				l.currentMatch = n
			}
		case tcell.KeyDown:
			if p := l.currentMatch.Prev(); p != nil {
				l.currentMatch = p
			}
		}
		return s
	}
	return ""
}

func (l *HistoryCmd) setAllMatch() {
	l.matchs.PushBackList(l.history)
	l.currentMatch = l.matchs.Front()
}
func (l *HistoryCmd) match() {
	if l.currentInputText == l.lastInputText {
		if l.matchs.Len() > 0 {
			return
		}
		l.setAllMatch()
	}
	l.matchs.Init()
	l.currentMatch = nil
	l.lastInputText = l.currentInputText
	if l.currentInputText == "" {
		l.setAllMatch()
		return
	}
	for e := l.history.Front(); e != nil; e.Next() {
		s := e.Value.(string)
		if strings.HasPrefix(strings.ToLower(s), strings.ToLower(l.currentInputText)) {
			l.matchs.PushBack(e.Value)
		}
	}
	if l.matchs.Len() > 0 {
		l.currentMatch = l.matchs.Front()
	}
}

// Add text into history record
func (l *HistoryCmd) Add(text string) {
	l.Lock()
	defer l.Unlock()

	// if exists, move to front
	for e := l.history.Front(); e != nil; e.Next() {
		if e.Value.(string) == text {
			l.history.MoveToFront(e)
			return
		}
	}

	// add new record
	if l.history.Len() >= l.maxLen {
		l.history.Remove(l.history.Back())
	}
	l.history.PushFront(text)
}

var historyCmd = NewHistoryCmd(historyCmdLength)

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
			historyCmd.Add(cmdstr)
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

	inputBox.SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
		key := e.Key()

		historyCmd.Lock()
		defer historyCmd.Unlock()
		switch key {
		case tcell.KeyUp, tcell.KeyDown:
			if !historyCmd.Scrolling {
				historyCmd.Scrolling = true
				historyCmd.SetCurrentText(inputBox.GetText())
			}
			s := historyCmd.NextMatch(key)
			if len(s) > 0 {
				inputBox.SetText(s)
			}
		default:
			if historyCmd.Scrolling {
				historyCmd.Scrolling = false
			}
		}

		return e
	})
}
