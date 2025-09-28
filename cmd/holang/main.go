package main

import (
	"flag"
	"internal/util/log"
	"os"
)

func main() {
	debug := flag.Bool("debug", false, "enable debug mode")
	flag.Parse()

	if *debug {
		log.EnableDebug()
	}

	args := flag.Args()
	if len(args) != 1 {
		log.Fatal("Usage: holang [--debug] [file]", log.A("args", os.Args))
		return
	}

	fileName := args[0]
	log.Info("Run HOLANG VM", log.S("file", fileName))

	bytes, err := readFile(fileName)
	if err != nil {
		log.Fatal("Read file error", log.S("file", fileName), log.E(err))
		return
	}

	err = run(bytes)
	if err != nil {
		log.Fatal("Run error", log.E(err))
		return
	}
}
