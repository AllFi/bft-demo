package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/AllFi/bft-demo/node"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init [numberOfNodes]",
	Short: "Initializes validator node directories",
	Long:  "Initializes validator node directories",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		numberOfNodes, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println(err)
			return
		}

		err = cleanUp(baseDir)
		if err != nil {
			fmt.Println(err)
			return
		}

		_, err = node.InitNewNodes(baseDir, numberOfNodes)
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}

func cleanUp(basePath string) (err error) {
	return os.RemoveAll(basePath)
}
