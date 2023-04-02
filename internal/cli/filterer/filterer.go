package filterer

import (
	"flag"
	"log"

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
	fs.StringVar(&cfg.GRPCAddr, "grpc-addr", "localhost:8001", "gRPC address")
	return subcommands.NewCommand(fs.Name(), fs, func() error {
		srv, err := server.New(cfg)
		if err != nil {
			return err
		}
		log.Println("filterer server started", "grpc-addr", cfg.GRPCAddr)
		return srv.Serve()
	})
}
