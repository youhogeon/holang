package main

import (
	"internal/util/log"
	"os"
)

func main() {
	if len(os.Args) > 2 {
		log.Fatal("Usage: lhox [file]", log.A("args", os.Args))

		return
	}

	if len(os.Args) == 2 {
		fileName := os.Args[1]
		log.Info("LHOX with file", log.S("file", fileName))

		runFile(fileName)

		return
	}

	log.Info("LHOX Loop Start")
	runLoop()
}
