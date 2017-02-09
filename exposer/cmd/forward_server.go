// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// forward_serverCmd represents the forward-server command
var forward_serverCmd = &cobra.Command{
	Use:   "forward-server",
	Short: "A server for forwarding network traffic",
	Long: `forward and forward-server are a pair for forwarding network traffic
via websocket protocal. So it can through 80 or 443 port via HTTP(s) protocal.`,
}

func init() {
	RootCmd.AddCommand(forward_serverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// forward_serverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// forward_serverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	var (
		addr     = "0.0.0.0:9000"
		key      = ""
		protocal = "ws"
	)
	forward_serverCmd.Flags().StringVarP(&addr, "addr", "a", addr, "listen address")
	forward_serverCmd.Flags().StringVarP(&key, "key", "k", key, "auth key")
	forward_serverCmd.Flags().StringVarP(&protocal, "protocal", "", protocal, "selected protocal")

	forward_serverCmd.Run = func(cmd *cobra.Command, args []string) {
		fmt.Println("key:", key)
	}
}
