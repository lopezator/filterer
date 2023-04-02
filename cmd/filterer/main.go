package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/lopezator/filterer/internal/cli/filterer"
	"github.com/peterbourgon/ff/v3"
)

func main() {
	filtererCmd := filterer.NewFiltererCommand()
	err := filtererCmd.Execute(os.Args[1:], func(fs *flag.FlagSet, args []string) error {
		return ff.Parse(fs, args,
			ff.WithConfigFileFlag("config"),
			ff.WithConfigFileParser(ff.PlainParser),
			ff.WithEnvVarPrefix(strings.ToUpper(filtererCmd.Name())),
		)
	})
	if err != nil {
		log.Fatal(err)
	}
}
