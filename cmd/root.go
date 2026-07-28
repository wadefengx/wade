package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "wade",
	Short: "All-in-one Node.js version & npm registry manager",
	Long: `Wade manages Node.js versions and npm/yarn/pnpm registries.
Single binary, installed once, no Node.js dependency.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolP("help", "h", false, "help for wade")
}