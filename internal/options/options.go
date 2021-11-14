package options

type Options struct {
	Watchables string `name:"watchables" help:"The file with your watchables." type:"existingfile" default:"watchables.yml"`
	GitName    string `name:"git-username" help:"The username to create a commit with." default:"Basechange"`
	GitEmail   string `name:"git-email" help:"The email to create a commit with." default:"basechange@github.com"`
}
