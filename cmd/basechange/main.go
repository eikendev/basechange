// Package main provides the main function as a starting point of this tool.
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/alecthomas/kong"

	ref "github.com/distribution/reference"
	log "github.com/sirupsen/logrus"

	"github.com/eikendev/basechange/internal/git"
	"github.com/eikendev/basechange/internal/options"
	"github.com/eikendev/basechange/internal/registry"
	"github.com/eikendev/basechange/internal/watchables"
)

var opts = options.Options{}

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp: true,
	})

	log.SetOutput(os.Stdout)

	log.SetLevel(log.InfoLevel)
}

func processImage(opts *options.Options, info watchables.Watchable) (string, error) {
	imageURL, err := ref.ParseNormalizedNamed(info.Image)
	if err != nil {
		return "", fmt.Errorf("failed to parse image name: %w", err)
	}

	log.Printf("Checking if %s has changed...", imageURL)

	digest, err := registry.GetImageDigest(info.Image)
	if err != nil {
		return "", err
	}

	if info.CachedDigest != digest {
		log.Printf("%s has changed!", imageURL)

		if err := git.Commit(opts, info.Repository, info.DeployKey, digest); err != nil {
			return "", err
		}

		return digest, nil
	}

	return "", nil
}

func processWatchables(opts *options.Options, ws *watchables.Watchables) error {
	if ws == nil {
		return errors.New("received nil watchables")
	}

	modified := false
	hasErrors := false

	for name, info := range *ws {
		newDigest, err := processImage(opts, info)
		if err != nil {
			log.Errorf("Error processing %s: %s", name, err)
			hasErrors = true
			continue
		}

		if newDigest != "" {
			info.CachedDigest = newDigest
			(*ws)[name] = info
			modified = true
		}
	}

	if modified {
		if err := watchables.Write(opts.Watchables, ws); err != nil {
			return fmt.Errorf("failed to write watchables: %w", err)
		}
	}

	if hasErrors {
		return errors.New("one or more images failed to process")
	}

	return nil
}

func main() {
	kong.Parse(
		&opts,
		kong.Description(fmt.Sprintf("%s (%s)", version, date)),
	)

	if opts.Debug {
		log.SetLevel(log.DebugLevel)
	}

	ws, err := watchables.Read(opts.Watchables)
	if err != nil {
		log.Fatal(err)
	}

	if err := processWatchables(&opts, ws); err != nil {
		log.Fatal(err)
	}
}
