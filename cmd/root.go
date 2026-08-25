package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "delay-kafka",
	Short: "A kafka delayed message transponder",
	Long:  `A kafka delayed message transponder`,
	RunE: func(_ *cobra.Command, _ []string) error {
		app, clean, err := wireApp()
		if err != nil {
			return fmt.Errorf("initialize application: %w", err)
		}
		signalCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		defer stop()

		eg, egCtx := errgroup.WithContext(signalCtx)
		eg.Go(func() error {
			return app.StartConsumer()
		})
		eg.Go(func() error {
			return app.StartTask()
		})
		eg.Go(func() error {
			<-egCtx.Done()
			clean()
			return nil
		})
		if err := eg.Wait(); err != nil {
			if signalCtx.Err() != nil {
				return nil
			}
			return fmt.Errorf("run application: %w", err)
		}
		return nil
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
