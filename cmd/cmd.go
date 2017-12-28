package cmd

import (
	"fmt"
	"os"
	"github.com/spf13/cobra"
	"github.com/Bestfeel/mou/markdown"
)

var (
	globalAddr = ":7070"
	globalPath = "."
)
var RootCmd = &cobra.Command{
	Use:   "Document",
	Short: "Powerful Document center",
	Long:  `This is a powerful online tool about Document center`,
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		markdown.RunMarkDownServer(globalAddr, globalPath)
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.Flags().StringVarP(&globalAddr, "addr", "a", ":7070", "server address")
	RootCmd.Flags().StringVarP(&globalPath, "path", "p", ".", "sever path")
}
