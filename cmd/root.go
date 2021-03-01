package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

var (
	baseDir string
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&baseDir, "baseDir", "", getDefaultBaseDir(), "directory to use for databases")
}

var rootCmd = &cobra.Command{
	Use:   "bft-demo",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("please specify an action")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func getDefaultBaseDir() string {
	home, err := homedir.Dir()
	if err != nil {
		panic(err)
	}

	return filepath.Join(home, ".bft-demo")
}
