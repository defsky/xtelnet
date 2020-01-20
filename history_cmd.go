package main

import (
	"bufio"
	"container/list"
	"os"
	"strings"
	"sync"

	"github.com/gdamore/tcell"
)

const cacheFileName = ".history"

type HistoryCmd struct {
	mu               sync.Mutex
	scrolling        bool
	maxLen           int
	currentInputText string
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
	if len(strings.Trim(text, " ")) <= 0 {
		text = ""
	}
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
	defer func() {
		l.currentMatch = l.matchs.Front()
		l.mu.Unlock()
	}()

	l.mu.Lock()

	l.matchs.Init()
	l.currentMatch = nil

	if l.currentInputText == "" {
		l.matchs.PushBackList(l.history)
		l.matchs.PushFront(" ")
		return
	}

	for e := l.history.Front(); e != nil; e = e.Next() {
		s := e.Value.(string)
		if s == l.currentInputText {
			continue
		}
		if strings.HasPrefix(strings.ToLower(s), strings.ToLower(l.currentInputText)) {
			l.matchs.PushBack(e.Value)
		}
	}
	if l.matchs.Len() > 0 {
		l.matchs.PushFront(l.currentInputText)
	}
}

// Add text into history record
func (l *HistoryCmd) Add(text string) {
	text = strings.TrimRight(text, " ")
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

func (l *HistoryCmd) Cache() {
	fd, err := os.Create(cacheFileName)
	if err != nil {
		return
	}
	defer fd.Close()

	for e := l.history.Back(); e != nil; e = e.Prev() {
		fd.WriteString(e.Value.(string) + "\n")
	}
}

func (l *HistoryCmd) LoadCache() {
	fd, err := os.Open(cacheFileName)
	if err != nil {
		return
	}
	defer fd.Close()

	rd := bufio.NewReader(fd)
	for line, err := rd.ReadString('\n'); err == nil; {
		line = strings.Trim(line, "\n ")
		l.history.PushFront(line)

		line, err = rd.ReadString('\n')
	}
}
