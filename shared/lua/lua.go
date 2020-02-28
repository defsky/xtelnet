package lua

import (
	"bufio"
	"os"
	"sync"

	lua "github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/parse"
)

// LStatePool is a pool for lua LState object
type LStatePool struct {
	m     sync.Mutex
	saved []*lua.LState
}

//Get a LState object
func (pl *LStatePool) Get() *lua.LState {
	pl.m.Lock()
	defer pl.m.Unlock()
	n := len(pl.saved)
	if n == 0 {
		return pl.New()
	}
	// x := pl.saved[n-1]
	// pl.saved = pl.saved[0 : n-1]
	x := pl.saved[0]
	pl.saved = pl.saved[1:]

	return x
}

// Put LState return to pool
func (pl *LStatePool) Put(L *lua.LState) {
	pl.m.Lock()
	defer pl.m.Unlock()
	pl.saved = append(pl.saved, L)
}

// New create a new LState object
func (pl *LStatePool) New() *lua.LState {
	L := lua.NewState()
	// setting the L up here.
	// load scripts, set global variables, share channels, etc...
	return L
}

// Shutdown close all LState object
func (pl *LStatePool) Shutdown() {
	for _, L := range pl.saved {
		L.Close()
	}
}

// Engine is a lua script engine
type Engine struct {
	pool *LStatePool
}

// NewEngine create a new Engine
func NewEngine() *Engine {
	return &Engine{
		pool: &LStatePool{
			saved: make([]*lua.LState, 0, 4),
		},
	}
}
func (e *Engine) Load(fname string) {

}

// Stop will stop the engine
func (e *Engine) Stop() {
	e.pool.Shutdown()
}

// Compile reads the passed lua file from disk and compiles it.
func (e *Engine) Compile(filePath string) (*lua.FunctionProto, error) {
	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(file)
	chunk, err := parse.Parse(reader, filePath)
	if err != nil {
		return nil, err
	}
	proto, err := lua.Compile(chunk, filePath)
	if err != nil {
		return nil, err
	}
	return proto, nil
}
