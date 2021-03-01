package cmd

import (
	"fmt"
	"strconv"

	"github.com/AllFi/bft-demo/node"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(recoverCmd)
}

var recoverCmd = &cobra.Command{
	Use:   "recover [nodeIndex] [referenceNodeIndex]",
	Short: "Recover",
	Long:  "Recover node by index (copies data of referenceNode)",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		nodeIndex, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println(err)
			return
		}

		refNodeIndex, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Println(err)
			return
		}

		err = node.Recover(baseDir, nodeIndex, refNodeIndex)
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}
