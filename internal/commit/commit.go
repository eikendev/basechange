package commit

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/eikendev/basechange/internal/options"
	"github.com/go-git/go-billy/v5/memfs"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
)

func Commit(opts *options.Options, url, deployKey, digest string) error {
	deployKeyBytes, err := base64.StdEncoding.DecodeString(deployKey)
	if err != nil {
		return err
	}

	auth, err := ssh.NewPublicKeys("git", deployKeyBytes, "")
	if err != nil {
		return err
	}

	fs := memfs.New()

	r, err := git.Clone(memory.NewStorage(), fs, &git.CloneOptions{
		Auth:  auth,
		URL:   url,
		Depth: 1,
	})
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	commitMsg := fmt.Sprintf("Updated base image: %s", digest)
	_, err = w.Commit(commitMsg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  opts.GitName,
			Email: opts.GitEmail,
			When:  time.Now(),
		},
	})
	if err != nil {
		return err
	}

	err = r.Push(&git.PushOptions{
		Auth: auth,
	})
	if err != nil {
		return err
	}

	return nil
}
