package main

import (
	"bufio"
	"fmt"
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
		log.Debug("LHOX with file", log.S("file", fileName))

		runFile(fileName)

		return
	}

	log.Info("LHOX Loop Start")
	runLoop()
}

func runFile(fileName string) {
	fileBody, err := os.ReadFile(fileName)
	if err != nil {
		log.Fatal("Read file error", log.S("file", fileName), log.E(err))
	}

	run(fileBody)
}

func runLoop() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("> ")
	for scanner.Scan() {
		line := scanner.Bytes()
		run(line)
		fmt.Print("> ")
	}

	if err := scanner.Err(); err != nil {
		log.Fatal("Scanner error", log.E(err))
	}
}

func run(source []byte) {

}
