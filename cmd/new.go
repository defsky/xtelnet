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
	"os"
	"os/exec"
	"syscall"

	"github.com/defsky/xtelnet/session"

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
	Run: func(c *cobra.Command, args []string) {
		sessionName := args[0]
		if os.Getppid() == 1 {
			s := session.NewSession(sessionName, cmdFile)
			s.Start()
			return
		}

		cmd := exec.Command(os.Args[0], os.Args[1:]...)
		cmd.Env = os.Environ()
		cmd.Stdin = nil
		cmd.Stdout = nil
		cmd.Stderr = nil
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setsid: true,
		}

		err := cmd.Start()
		if err == nil {

			fullSessionName := fmt.Sprintf("%d.%s", cmd.Process.Pid, sessionName)
			fmt.Printf("  %s	[Detached]\n", fullSessionName)

			cmd.Process.Release()
			os.Exit(0)

			// if isDetached {
			// 	fmt.Printf("  %s	[Detached]\n", fullSessionName)
			// 	os.Exit(0)
			// } else {
			// 	// xui.NewXUI().Attach(fullSessionName)
			// 	time.Sleep(time.Second)
			// 	cmd := exec.Command(os.Args[0], []string{"attach", fullSessionName}...)
			// 	cmd.Env = append(os.Environ(), "RUNEWIDTH_EASTASIAN=1")
			// 	err := cmd.Run()
			// 	if err != nil {
			// 		fmt.Println(err)
			// 		os.Exit(-1)
			// 	}
			// }
		}

		// pid, _, err := syscall.Syscall(syscall.SYS_FORK, 0, 0, 0)
		// if err != 0 {
		// 	fmt.Printf("%v\n", err)
		// 	return
		// }

		// if pid == 0 {
		// 	/*
		// 		// start child process
		// 		// Change the file mode mask
		// 		// _ = syscall.Umask(0)

		// 		// // create a new SID for the child process
		// 		// s_ret, s_errno := syscall.Setsid()
		// 		// if s_errno != nil {
		// 		// 	log.Printf("Error: syscall.Setsid errno: %d", s_errno)
		// 		// }
		// 		// if s_ret < 0 {
		// 		// 	return
		// 		// }

		// 		// if nochdir == 0 {
		// 		// 	os.Chdir("/")
		// 		// }

		// 		// if noclose == 0 {
		// 		// 	f, e := os.OpenFile("/dev/null", os.O_RDWR, 0)
		// 		// 	if e == nil {
		// 		// 		fd := f.Fd()
		// 		// 		syscall.Dup2(int(fd), int(os.Stdin.Fd()))
		// 		// 		syscall.Dup2(int(fd), int(os.Stdout.Fd()))
		// 		// 		syscall.Dup2(int(fd), int(os.Stderr.Fd()))
		// 		// 	}
		// 		// }
		// 	*/
		// 	ret, err := syscall.Setsid()
		// 	if err != nil || ret < 0 {
		// 		// return fmt.Errorf("fail to call setsid")
		// 		return
		// 	}

		// 	signal.Ignore(syscall.SIGHUP)
		// 	syscall.Umask(0)

		// 	name := args[0]
		// 	session.Create(name, cmdFile)
		// } else {

		// 	if isDetached {
		// 		fmt.Println("start session in detached mode")
		// 	}

		// 	sessionName := fmt.Sprintf("%d.%s", pid, args[0])

		// 	fmt.Println("Session name: ", sessionName)
		// }
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
	// newCmd.PersistentFlags().String("foo", "", "A help for foo")
	// newCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	newCmd.Flags().BoolVarP(&isDetached, "detach", "d", false, "create new session in detached status")
	newCmd.Flags().StringVarP(&cmdFile, "file", "f", "", "specify startup command file name")
}
