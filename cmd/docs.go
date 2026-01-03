package cmd

import (
	"fmt"
	"log"

	"github.com/phillip-england/thispage/pkg/docs"
	"github.com/spf13/cobra"
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Start a documentation server on port 8080",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting thispage documentation server on http://localhost:8080")
		if err := docs.Serve(); err != nil {
			log.Fatalf("Error starting docs server: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(docsCmd)
}
