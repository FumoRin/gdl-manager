package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fumorin/gdl-manager/internal/downloader"
	"github.com/spf13/cobra"
)

type App struct {
	Manager *downloader.DownloadManager
}

var app App

var rootCmd = &cobra.Command{
	Use:   "gdl",
	Short: "Download file from the internet",
	Long:  "A CLI Tool for downloading file from the internet",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		dir, err := os.UserConfigDir()
		var dbPath string
		if err == nil {
			gdlDir := filepath.Join(dir, "gdl")
			_ = os.MkdirAll(gdlDir, 0755)
			dbPath = filepath.Join(gdlDir, "gdl.db")
		} else {
			dbPath = "gdl.db"
		}
		repo, err := downloader.NewSQLiteRepository(dbPath)
		if err != nil {
			return fmt.Errorf("failed to initialized downloader database: %w", err)
		}

		app.Manager = downloader.NewDownloadManager(repo)
		app.Manager.StartWorker(3)

		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}
