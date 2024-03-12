package filterer

import (
	"flag"

	"github.com/glerchundi/subcommands"
	"github.com/lopezator/filterer/internal/server"
)

// NewFiltererCommand create and returns the root cli command.
func NewFiltererCommand() *subcommands.Command {
	filtererCmd := subcommands.NewCommand("filterer", flag.CommandLine, nil)
	filtererCmd.AddCommand(newServeCommand())
	return filtererCmd
}

// newServeCommand creates a new serve command and runs the server.
func newServeCommand() *subcommands.Command {
	cfg := &server.Config{}
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	_ = fs.String("config", "", "config file (optional)")
	fs.StringVar(&cfg.Addr, "addr", ":8001", "server address")
	return subcommands.NewCommand(fs.Name(), fs, func() error {
		srv, err := server.New(cfg)
		if err != nil {
			return err
		}
		return srv.Serve()
	})
}
