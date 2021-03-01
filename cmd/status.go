package cmd

import (
	"fmt"
	"strconv"

	"github.com/AllFi/bft-demo/node"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "changeStatus [nodeIndex] [status <- Correct,Malicious]",
	Short: "Changes node status by index",
	Long:  "Changes node status by index",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		nodeIndex, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println(err)
			return
		}

		status := args[1]

		resp, err := Send(fmt.Sprintf("http://localhost:%v/abci_query?path=\"status\"&data=\"%v\"", node.ShiftPort(26657, nodeIndex), status))
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(resp)
	},
}
