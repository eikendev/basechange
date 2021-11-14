package watchables

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type Watchable struct {
	Image        string `yaml:"image"`
	Repository   string `yaml:"repository"`
	DeployKey    string `yaml:"deploy_key"`
	CachedDigest string `yaml:"cached_digest"`
}

type Watchables map[string]Watchable

func Read(path string) (*Watchables, error) {
	yfile, err := ioutil.ReadFile(path)
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

func Write(path string, watchables *Watchables) error {
	out, err := yaml.Marshal(watchables)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, out, 0644)
	if err != nil {
		return err
	}

	return nil
}
