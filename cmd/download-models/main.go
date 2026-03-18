// download-models downloads model files for bundling into release archives.
// This tool has no CGO dependencies and can be built with CGO_ENABLED=0.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"

	parakeetmodel "github.com/dharmab/skyeye/pkg/recognizer/parakeet/model"
	pocketmodel "github.com/dharmab/skyeye/pkg/synthesizer/pocket/model"
)

func main() {
	dir := flag.String("dir", "models", "base directory to download model files into")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	parakeetDir := filepath.Join(*dir, parakeetmodel.DirName)
	if err := parakeetmodel.Verify(parakeetDir); err != nil {
		log.Printf("Parakeet model needs download: %v", err)
		if err := parakeetmodel.Download(ctx, parakeetDir); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Println("Parakeet model already present and verified")
	}

	pocketDir := filepath.Join(*dir, pocketmodel.DirName)
	if err := pocketmodel.Verify(pocketDir); err != nil {
		log.Printf("Pocket TTS model needs download: %v", err)
		if err := pocketmodel.Download(ctx, pocketDir); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Println("Pocket TTS model already present and verified")
	}
}
