package main

import (
	"flag"
	"internal/util/log"
	"os"
)

func main() {
	debugMode := flag.Bool("debug", false, "enable debug mode")
	runMode := flag.Bool("run", false, "enable run mode")
	flag.Parse()

	if *debugMode {
		log.EnableDebug()
	}

	args := flag.Args()
	if len(args) != 1 {
		log.Fatal("Usage: holangc [--debug] [--run] [file]", log.A("args", os.Args))
		return
	}

	fileName := args[0]
	log.Info("Run HOLANG compiler", log.S("file", fileName), log.B("run mode", *runMode), log.B("debug mode", *debugMode))

	baseName, source, err := readSource(fileName)
	if err != nil {
		log.Fatal("Read source error", log.S("file", fileName), log.E(err))
		return
	}
	compiled := compile(source)

	if *runMode {
		err := run(compiled)

		if err != nil {
			log.Fatal("Run error", log.E(err))
			return
		}
	} else {
		savedFileName, err := saveBytes(baseName, compiled)

		if err != nil {
			log.Fatal("Write output file error", log.S("file", savedFileName), log.E(err))
			return
		}

		log.Info("Saved output file", log.S("file", savedFileName))
	}

	// log.Info("HOLANG Loop Start", log.S("out", output))
	// runLoop()
}
