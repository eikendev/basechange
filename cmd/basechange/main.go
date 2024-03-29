// Package main provides the main function as a starting point of this tool.
package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/eikendev/basechange/internal/git"
	"github.com/eikendev/basechange/internal/options"
	"github.com/eikendev/basechange/internal/registry"
	"github.com/eikendev/basechange/internal/watchables"

	ref "github.com/distribution/reference"
	log "github.com/sirupsen/logrus"
)

var opts = options.Options{}

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp: true,
	})

	log.SetOutput(os.Stdout)

	log.SetLevel(log.InfoLevel)
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

	success := true

	for name, info := range *ws {
		imageURL, err := ref.ParseNormalizedNamed(info.Image)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("Checking if %s has changed...", imageURL)

		digest, err := registry.GetImageDigest(info.Image)
		if err != nil {
			log.Errorf("%s", err)
			success = false
			continue
		}

		if info.CachedDigest != digest {
			log.Printf("%s has changed!", imageURL)

			err := git.Commit(&opts, info.Repository, info.DeployKey, digest)
			if err != nil {
				log.Errorf("%s", err)
				success = false
				continue
			}

			// Set digest after commit, so we can retry if the commit fails.
			info.CachedDigest = digest
			(*ws)[name] = info
		}
	}

	err = watchables.Write(opts.Watchables, ws)
	if err != nil {
		log.Fatal(err)
	}

	if !success {
		os.Exit(1)
	}
}
