package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/eikendev/basechange/internal/git"
	"github.com/eikendev/basechange/internal/options"
	"github.com/eikendev/basechange/internal/registry"
	"github.com/eikendev/basechange/internal/watchables"

	ref "github.com/docker/distribution/reference"
	log "github.com/sirupsen/logrus"
)

var (
	opts = options.Options{}
)

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

			info.CachedDigest = digest
			(*ws)[name] = info

			err := git.Commit(&opts, info.Repository, info.DeployKey, digest)
			if err != nil {
				log.Errorf("%s", err)
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
