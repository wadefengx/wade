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
		// wade -i shortcut
		if skipAll, _ := cmd.Flags().GetBool("init"); skipAll {
			runInteractiveWizard(cmd, args)
			os.Exit(0)
		}
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
	rootCmd.Flags().BoolP("init", "i", false, "Run interactive setup wizard (alias: wade init)")

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Interactive setup wizard for wade",
		RunE:  runInteractiveWizard,
	}
	initCmd.Flags().BoolP("yes", "y", false, "Skip prompts, use defaults")
	rootCmd.AddCommand(initCmd)
}