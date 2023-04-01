// Package options defines the global options of this tool.
package options

// Options represents the global options of this tool.
type Options struct {
	Watchables string `name:"watchables" help:"The file with your watchables." type:"existingfile" default:"watchables.yml" env:"BASECHANGE_WATCHABLES"`
	GitName    string `name:"git-name" help:"The name to create a commit with." default:"Basechange" env:"BASECHANGE_GIT_NAME"`
	GitEmail   string `name:"git-email" help:"The email to create a commit with." default:"basechange@github.com" env:"BASECHANGE_GIT_EMAIL"`
	Debug      bool   `name:"debug" help:"Print debug messages." default:"false"`
}
