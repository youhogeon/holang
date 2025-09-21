package main

import (
	"bytes"
	"compress/gzip"
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

	outputFileName := strings.TrimSuffix(fileName, ".holang") + ".hbc"

	fileBody, err := os.ReadFile(fileName)
	if err != nil {
		fileBody, err = os.ReadFile(fileName + ".holang")
		if err != nil {
			log.Fatal("Read file error", log.S("file", fileName), log.E(err))
		}
	}

	log.Info("Compile Start", log.S("file", fileName))
	result := compile(fileBody)

	// comporessed, err := compress(result)
	// if err != nil {
	// 	log.Fatal("Compress error", log.E(err))
	// }

	err = os.WriteFile(outputFileName, result, 0644)
	if err != nil {
		log.Fatal("Write file error", log.S("file", outputFileName), log.E(err))
	}

	log.Info("Compile Success", log.S("output", outputFileName))
}

func compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, err := gz.Write(data)
	if err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// func decompress(data []byte) ([]byte, error) {
// 	buf := bytes.NewReader(data)
// 	gz, err := gzip.NewReader(buf)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer gz.Close()
// 	return io.ReadAll(gz)
// }
