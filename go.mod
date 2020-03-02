module github.com/defsky/xtelnet

go 1.13

require (
	github.com/gdamore/tcell v1.3.0
	github.com/rivo/tview v0.0.0-20200219210816-cd38d7432498
	github.com/spf13/cobra v0.0.6
	github.com/yuin/gopher-lua v0.0.0-20191220021717-ab39c6098bdb
	golang.org/x/text v0.3.2
)

replace github.com/rivo/tview => ./tview

replace github.com/defsky/xtelnet => ./
