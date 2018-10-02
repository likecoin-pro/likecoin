package config

import (
	"flag"
	"fmt"
	"os"
)

const (
	Version         = "0.1.002"
	ApplicationName = "likecd"
)

func ParseArgs() {
	var (
		argHelp    = flag.Bool("help", false, "Show this help")
		argVersion = flag.Bool("version", false, "Show software version")
	)
	flag.Parse()

	if *argHelp {
		flag.PrintDefaults()
		os.Exit(0)
		return
	}
	if *argVersion {
		fmt.Printf("%s (%s) version %s\n", ApplicationName, ProgramType, Version)
		os.Exit(0)
		return
	}
}
