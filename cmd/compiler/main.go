package main

import (
	"internal/util/log"
	"os"
	"strings"
)

func main() {
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

	if len(filtered) != 1 {
		log.Fatal("Usage: holang [--debug] [file]", log.A("args", os.Args))
		return
	}

	fileName = filtered[0]

	var outputFileName string
	if strings.HasSuffix(outputFileName, ".holang") {
		outputFileName = strings.TrimSuffix(outputFileName, ".holang") + ".hbc"
	} else {
		outputFileName = outputFileName + ".hbc"
	}

	fileBody, err := os.ReadFile(fileName)
	if err != nil {
		fileBody, err = os.ReadFile(fileName + ".holang")
		if err != nil {
			log.Fatal("Read file error", log.S("file", fileName), log.E(err))
		}
	}

	log.Info("Compile Start", log.S("file", fileName))
	result := compile(fileBody)

	err = os.WriteFile(outputFileName, result, 0644)
	if err != nil {
		log.Fatal("Write file error", log.S("file", outputFileName), log.E(err))
	}

	log.Info("Compile Success", log.S("output", outputFileName))
}
