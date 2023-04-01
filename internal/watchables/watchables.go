// Package watchables provides functionality for the watchables file.
package watchables

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Watchable contains information of an image to be watched.
type Watchable struct {
	Image        string `yaml:"image"`
	Repository   string `yaml:"repository"`
	DeployKey    string `yaml:"deploy_key"`
	CachedDigest string `yaml:"cached_digest"`
}

// Watchables is a collection of Watchable objects.
type Watchables map[string]Watchable

// Read returns the contents of the watchables file.
func Read(path string) (*Watchables, error) {
	yfile, err := os.ReadFile(path) //#nosec G304
	if err != nil {
		return nil, err
	}

	var watchables Watchables

	err = yaml.Unmarshal(yfile, &watchables)
	if err != nil {
		return nil, err
	}

	for name, info := range watchables {
		if info.Image == "" {
			return nil, fmt.Errorf("image may not be empty for watchable %s", name)
		}
		if info.Repository == "" {
			return nil, fmt.Errorf("repository may not be empty for watchable %s", name)
		}
		if info.DeployKey == "" {
			return nil, fmt.Errorf("deploy_key may not be empty for watchable %s", name)
		}
	}

	return &watchables, nil
}

// Write writes the provided watchables object to the watchables file.
func Write(path string, watchables *Watchables) error {
	out, err := yaml.Marshal(watchables)
	if err != nil {
		return err
	}

	err = os.WriteFile(path, out, 0o600)
	if err != nil {
		return err
	}

	return nil
}
