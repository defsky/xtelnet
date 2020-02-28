/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"syscall"
	"xtelnet/session"

	"github.com/spf13/cobra"
)

var isDetached bool
var cmdFile string

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new <session name>",
	Short: "create a new session",
	Long:  `create a new xtelnet session`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("main start")

		pid, _, err := syscall.Syscall(syscall.SYS_FORK, 0, 0, 0)
		if err != 0 {
			fmt.Printf("%v\n", err)
			return
		}

		if pid == 0 {
			// start child process
			name := args[0]
			session.Create(name, cmdFile)
		} else {

			if isDetached {
				fmt.Println("start session in detached mode")
			}

			sessionName := fmt.Sprintf("%d.%s", pid, args[0])

			fmt.Println("Session name: ", sessionName)
		}
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
	// newCmd.PersistentFlags().String("foo", "", "A help for foo")
	// newCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	newCmd.Flags().BoolVarP(&isDetached, "detach", "d", false, "create new session in detached status")
	newCmd.Flags().StringVarP(&cmdFile, "file", "f", "", "specify startup command file name")
}
