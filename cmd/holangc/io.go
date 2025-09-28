package main

import (
	_ "embed"
	"fmt"
	"internal/runtime"
	"internal/serialize"
	"os"
	"strings"
)

const EXT_SOURCE = ".holang"
const EXT_BYTES = ".hbc"

//go:embed VERSION
var VERSION string

func readSource(fileName string) (string, []byte, error) {
	if strings.HasSuffix(fileName, EXT_SOURCE) {
		fileBody, err := os.ReadFile(fileName)
		if err != nil {
			return "", nil, fmt.Errorf("failed to read file %s: %w", fileName, err)
		}

		baseName := strings.TrimSuffix(fileName, EXT_SOURCE)
		return baseName, fileBody, nil
	}

	fileBody, err := os.ReadFile(fileName)
	if err == nil {
		return fileName, fileBody, nil
	}

	altName := fileName + EXT_SOURCE
	fileBody, err = os.ReadFile(altName)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read file %s or %s: %w", fileName, altName, err)
	}

	return fileName, fileBody, nil
}

func saveBytes(baseName string, fn *runtime.ObjFunction) (string, error) {
	fileName := baseName + EXT_BYTES

	bytes, err := serialize.Serialize(fn, VERSION)
	if err != nil {
		return "", fmt.Errorf("failed to serialize: %w", err)
	}

	err = os.WriteFile(fileName, bytes, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write file %s: %w", fileName, err)
	}

	return fileName, nil
}
