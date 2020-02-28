package shared

import (
	"container/list"
	"sync"
)

type ScriptEngine interface {
	Stop()
}

type Queue struct {
	l *list.List
	m sync.Mutex
}

func NewQueue() *Queue {
	return &Queue{l: list.New()}
}

func (q *Queue) PushBack(v interface{}) {
	if v == nil {
		return
	}
	q.m.Lock()
	defer q.m.Unlock()
	q.l.PushBack(v)
}

func (q *Queue) PushBackQueue(other *Queue) {
	if other == nil {
		return
	}
	q.m.Lock()
	defer q.m.Unlock()
	q.l.PushBackList(other.l)
}
func (q *Queue) PushFront(v interface{}) {
	if v == nil {
		return
	}
	q.m.Lock()
	defer q.m.Unlock()
	q.l.PushFront(v)
}
func (q *Queue) PushFrontQueue(other *Queue) {
	if other == nil {
		return
	}
	q.m.Lock()
	defer q.m.Unlock()
	q.l.PushFrontList(other.l)
}
func (q *Queue) Front() *list.Element {
	q.m.Lock()
	defer q.m.Unlock()
	return q.l.Front()
}

func (q *Queue) Back() *list.Element {
	q.m.Lock()
	defer q.m.Unlock()
	return q.l.Back()
}

func (q *Queue) Remove(e *list.Element) {
	if e == nil {
		return
	}
	q.m.Lock()
	defer q.m.Unlock()
	q.l.Remove(e)
}

func (q *Queue) Len() int {
	q.m.Lock()
	defer q.m.Unlock()
	return q.l.Len()

}
