package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "delay-kafka",
	Short: "A kafka delayed message transponder",
	Long:  `A kafka delayed message transponder`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		app, clean, err := wireApp()
		if err != nil {
			panic(err)
		}
		safeShutdown(clean)
		return app.Start()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func safeShutdown(cleanup func()) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	go func() {
		for {
			sig := <-ch
			fmt.Printf("Got %s signal. Aborting...\n", sig)
			cleanup()
			os.Exit(1)
		}
	}()
}
