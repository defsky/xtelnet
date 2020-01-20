package main

import (
	"container/list"
	"strings"
	"sync"

	"github.com/gdamore/tcell"
)

const historyCmdLength = 1000

type HistoryCmd struct {
	mu               sync.Mutex
	scrolling        bool
	maxLen           int
	currentInputText string
	lastInputText    string
	history          *list.List
	matchs           *list.List
	currentMatch     *list.Element
}

func NewHistoryCmd(len int) *HistoryCmd {

	return &HistoryCmd{
		maxLen:  historyCmdLength,
		history: list.New(),
		matchs:  list.New(),
	}
}
func (l *HistoryCmd) IsScrolling() bool {
	return l.scrolling
}
func (l *HistoryCmd) SetScrolling(b bool) {
	l.scrolling = b
}
func (l *HistoryCmd) SetCurrentText(text string) {
	l.currentInputText = text
}

func (l *HistoryCmd) NextMatch(key tcell.Key) string {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.currentMatch == nil {
		return ""
	}

	switch key {
	case tcell.KeyUp:
		if n := l.currentMatch.Next(); n != nil {
			l.currentMatch = n
		}
	case tcell.KeyDown:
		if p := l.currentMatch.Prev(); p != nil {
			l.currentMatch = l.currentMatch.Prev()
		}
	}

	return l.currentMatch.Value.(string)
}

func (l *HistoryCmd) Match() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.matchs.Init()
	l.currentMatch = nil
	l.lastInputText = l.currentInputText
	if l.currentInputText == "" {
		l.matchs.PushBackList(l.history)
		l.matchs.PushFront("")
		l.currentMatch = l.matchs.Front()
		return
	}

	l.history.PushFront(l.currentInputText)
	for e := l.history.Front(); e != nil; e = e.Next() {
		s := e.Value.(string)
		if strings.HasPrefix(strings.ToLower(s), strings.ToLower(l.currentInputText)) {
			l.matchs.PushBack(e.Value)
		}
	}

	l.currentMatch = l.matchs.Front()
}

// Add text into history record
func (l *HistoryCmd) Add(text string) {
	if len(text) <= 0 {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	// if exists, move to front
	for e := l.history.Front(); e != nil; e = e.Next() {
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
