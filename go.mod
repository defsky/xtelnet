module xtelnet

go 1.13

require (
	github.com/gdamore/tcell v1.3.0
	github.com/rivo/tview v0.0.0-20191231100700-c6236f442139
	github.com/spf13/cobra v0.0.5
	github.com/takama/daemon v0.11.0
	github.com/yuin/gopher-lua v0.0.0-20191220021717-ab39c6098bdb
	golang.org/x/text v0.3.2
)

replace github.com/rivo/tview => ./tview
