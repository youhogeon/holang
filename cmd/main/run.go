package main

import (
	"bufio"
	"fmt"
	"internal/bytecode"
	"internal/util/log"
	"os"
)

func runFile(fileName string) {
	fileBody, err := os.ReadFile(fileName)
	if err != nil {
		fileBody, err = os.ReadFile(fileName + ".holang")
		if err != nil {
			log.Fatal("Read file error", log.S("file", fileName), log.E(err))
		}
	}

	run(fileBody)
}

func runLoop() {
	inputScanner := bufio.NewScanner(os.Stdin)

	log.StdOut("> ")
	for inputScanner.Scan() {
		line := inputScanner.Bytes()
		run(line)
		log.StdOut("> ")
	}

	if err := inputScanner.Err(); err != nil {
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

	test()
}

func test() {
	c := bytecode.NewChunk()

	c.Write(123, bytecode.OP_RETURN)
	c.Write(123, bytecode.OP_COSNTANT, c.AddConstant(500))
	c.Write(123, bytecode.OP_COSNTANT, c.AddConstant("HOLang2"))
	c.Write(123, bytecode.OP_RETURN)
	c.Disassemble()

}
