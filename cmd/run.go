package cmd

import (
	"fmt"
	"strconv"

	"github.com/AllFi/bft-demo/app"
	"github.com/AllFi/bft-demo/node"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run [nodeIndex]",
	Short: "Run",
	Long:  "Run node by index",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		nodeIndex, err := strconv.Atoi(args[0])
		if err != nil {
			return
		}

		app := app.NewApplication()
		err = node.Run(app, baseDir, nodeIndex)
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}
