package main

import (
	"fmt"
	"internal/serialize"
	"internal/util/log"
	"internal/vm"
	"os"
	"strings"
)

const EXT_BYTES = ".hbc"

var SUPPORTS_VERSION = []string{
	"2.0.0",
}

func readFile(fileName string) ([]byte, error) {
	if strings.HasSuffix(fileName, EXT_BYTES) {
		fileBody, err := os.ReadFile(fileName)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", fileName, err)
		}

		return fileBody, nil
	}

	fileBody, err := os.ReadFile(fileName)
	if err == nil {
		return fileBody, nil
	}

	altName := fileName + EXT_BYTES
	fileBody, err = os.ReadFile(altName)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s or %s: %w", fileName, altName, err)
	}

	return fileBody, nil
}

func run(bytes []byte) error {
	program, err := serialize.Deserialize(bytes, SUPPORTS_VERSION)
	if err != nil {
		return fmt.Errorf("failed to deserialize: %w", err)
	}

	machine := vm.NewVM()
	result := machine.Run(program.RootFunction)

	log.Info("VM interpret finished", log.A("result", result))

	return nil
}
