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
	"net"
	"os"
	"path/filepath"

	"github.com/defsky/xtelnet/proto"
	"github.com/defsky/xtelnet/session"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list xtelnet sessions",
	Long:  `list xtelnet sessions`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {

		homedir, err := session.SocketHomeDir()
		if err != nil {
			fmt.Println(err)
		}

		sessions, err := session.GetSessionList(homedir)
		if err != nil {
			fmt.Print(err)
		}

		if len(sessions) > 0 {
			count := 0
			fmt.Println("There are sessions on:")
			for _, v := range sessions {
				s := getSessionStatus(homedir, v)
				if len(s) > 0 {
					fmt.Printf("    %s\n", s)
					count++
				}
			}
			fmt.Printf(" %d Sockets in %s\n\n", count, homedir)
		} else {
			fmt.Println("There is no session\n")
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func getSessionStatus(d, s string) string {
	fpath := filepath.Join(d, s)
	addr, err := net.ResolveUnixAddr("unix", fpath)
	if err != nil {
		os.Remove(fpath)
		return ""
	}

	conn, err := net.DialUnix("unix", nil, addr)
	if err != nil {
		os.Remove(fpath)
		return ""
	}
	defer conn.Close()

	p := &proto.Packet{}
	p.Opcode = proto.CM_QUERY_DETACH_STATUS
	err = proto.WritePacket(conn, p)
	if err != nil {
		os.Remove(fpath)
		return ""
	}
	retp, err := proto.ReadPacket(conn)
	if err != nil {
		os.Remove(fpath)
		return ""
	}
	switch retp.Opcode {
	case proto.SM_DETACH_STATUS:
		b, _ := retp.ReadByte()
		if uint8(b) == 0 {
			return fmt.Sprintf("%-20s  (%s)", s, "Attached")
		}
		if uint8(b) == 1 {
			return fmt.Sprintf("%-20s  (%s)", s, "Detached")
		}
		return fmt.Sprintf("%-20s  (%s)", s, "Unknown ")
	default:
		return fmt.Sprintf("%-20s  (%s)", s, "Unknown ")
	}
}
