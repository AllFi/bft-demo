package cmd

import (
	"fmt"
	"strconv"

	"github.com/AllFi/bft-demo/node"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(txCmd)
}

var txCmd = &cobra.Command{
	Use:   "tx [nodeIndex] [value]",
	Short: "Broadcasts transaction",
	Long:  "Broadcasts transaction",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		nodeIndex, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println(err)
			return
		}

		value, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Println(err)
			return
		}

		resp, err := Send(fmt.Sprintf("http://localhost:%v/broadcast_tx_commit?tx=\"%v\"", node.ShiftPort(26657, nodeIndex), value))
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(resp)
	},
}
