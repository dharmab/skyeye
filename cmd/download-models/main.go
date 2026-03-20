// download-models downloads model files for bundling into release archives.
// This tool has no CGO dependencies and can be built with CGO_ENABLED=0.
package main

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/dharmab/skyeye/pkg/models"
	parakeet "github.com/dharmab/skyeye/pkg/recognizer/parakeet/model"
	pocket "github.com/dharmab/skyeye/pkg/synthesizer/pocket/model"
)

var dir string

func init() {
	rootCmd.Flags().StringVar(&dir, "dir", "models", "Base directory to download model files into")
}

var rootCmd = &cobra.Command{
	Use:   "download-models",
	Short: "Download model files for bundling into release archives",
	Run: func(_ *cobra.Command, _ []string) {
		run()
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("download-models exited with error")
	}
}

func run() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if err := models.Setup(ctx, "parakeet", filepath.Join(dir, parakeet.DirName), parakeet.Verify, parakeet.Download); err != nil {
		log.Fatal().Err(err).Msg("failed to set up parakeet model")
	}
	if err := models.Setup(ctx, "pocket-tts", filepath.Join(dir, pocket.DirName), pocket.Verify, pocket.Download); err != nil {
		log.Fatal().Err(err).Msg("failed to set up pocket-tts model")
	}
}
