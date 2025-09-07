package main

import (
	"internal/util/log"
	"os"
)

func main() {
	// Simple arg parsing: --debug optional + optional file
	args := os.Args[1:]
	var fileName string

	filtered := make([]string, 0, len(args))
	for _, a := range args {
		if a == "--debug" {
			log.EnableDebug()
			continue
		}
		filtered = append(filtered, a)
	}

	if len(filtered) > 1 { // too many non-flag args
		log.Fatal("Usage: holang [--debug] [file]", log.A("args", os.Args))
		return
	}

	if len(filtered) == 1 {
		fileName = filtered[0]
		log.Info("HOLANG with file", log.S("file", fileName))
		runFile(fileName)
		return
	}

	log.Info("HOLANG Loop Start")
	runLoop()
}
