// download-models downloads Parakeet TDT model files for bundling into release archives.
// This tool has no CGO dependencies and can be built with CGO_ENABLED=0.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/dharmab/skyeye/pkg/recognizer/parakeet/model"
)

func main() {
	dir := flag.String("dir", filepath.Join("models", model.DirName), "directory to download model files into")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := model.Download(ctx, *dir); err != nil {
		log.Fatal(err)
	}
}
