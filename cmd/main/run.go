package main

import (
	"bufio"
	"fmt"
	"internal/scanner"
	"internal/util/log"
	"os"
)

func runFile(fileName string) {
	fileBody, err := os.ReadFile(fileName)
	if err != nil {
		log.Fatal("Read file error", log.S("file", fileName), log.E(err))
	}

	run(fileBody)
}

func runLoop() {
	scanner := bufio.NewScanner(os.Stdin)

	log.StdOut("> ")
	for scanner.Scan() {
		line := scanner.Bytes()
		run(line)
		log.StdOut("> ")
	}

	if err := scanner.Err(); err != nil {
		log.Fatal("Scanner error", log.E(err))
	}
}

func run(source []byte) {
	sourceStr := string(source)

	log.InfoIfEnabled("Run source", func() []log.Field {
		_sourceStr := sourceStr

		if len(source) > 100 {
			_sourceStr = string(source[:100]) + "...(more " + fmt.Sprint(len(source)) + " bytes)"
		}

		return []log.Field{log.S("source", _sourceStr)}
	})

	scanner := scanner.NewScanner(sourceStr)
	tokens, errs := scanner.ScanTokens()

	log.Debug("Scan complete", log.A("tokens", tokens), log.A("errors", errs))
}
