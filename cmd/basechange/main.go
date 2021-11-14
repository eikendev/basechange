package main

import (
	"log"
	"os"

	"github.com/alecthomas/kong"
	"github.com/eikendev/basechange/internal/commit"
	"github.com/eikendev/basechange/internal/options"
	"github.com/eikendev/basechange/internal/registry"
	"github.com/eikendev/basechange/internal/watchables"

	ref "github.com/docker/distribution/reference"
)

var opts options.Options

func main() {
	kong.Parse(&opts)

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
			log.Printf("%s", err)
			success = false
			continue
		}

		if info.CachedDigest != digest {
			log.Printf("%s has changed!", imageURL)

			info.CachedDigest = digest
			(*ws)[name] = info

			err := commit.Commit(&opts, info.Repository, info.DeployKey, digest)
			if err != nil {
				log.Printf("%s", err)
				success = false
				continue
			}
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
