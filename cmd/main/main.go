package main

import (
	"flag"
	"internal/util/log"
	"os"
)

func main() {
	debug := flag.Bool("debug", false, "enable debug mode")
	output := *flag.String("o", "", "output file")
	flag.Parse()

	if *debug {
		log.EnableDebug()
	}

	args := flag.Args()
	if len(args) > 1 {
		log.Fatal("Usage: holang [--debug] [file]", log.A("args", os.Args))
		return
	}

	if len(args) == 1 {
		fileName := args[0]
		log.Info("HOLANG with file", log.S("file", fileName), log.S("out", output))
		runFile(fileName)
		return
	}

	log.Info("HOLANG Loop Start", log.S("out", output))
	runLoop()
}
