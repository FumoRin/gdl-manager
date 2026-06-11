package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use: "gdl",
	Short: "Download file from the internet",
	Long: "A CLI Tool for downloading file from the internet",
}

func Execute() error {
	return rootCmd.Execute()
}
