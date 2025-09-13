package main

import (
	"bufio"
	"fmt"
	"internal/ast"
	"internal/bytecode"
	"internal/parser"
	"internal/scanner"
	"internal/util/log"
	"internal/vm"
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

	scanner := scanner.NewScanner(sourceStr)
	tokens, errs := scanner.ScanTokens()

	log.Debug("Scan complete", log.A("tokens", tokens), log.A("errors", errs))

	if len(errs) > 0 {
		return
	}

	p := parser.NewParser(tokens)
	printer := ast.NewAstPrinter()

	statements, errs := p.Parse()

	log.Debug("Parse complete", log.A("ast", statements), log.A("errors", errs))

	for _, stmt := range statements {
		log.Debug("AST", log.S("astStr", printer.PrintStmt(stmt)))
	}

	if len(errs) > 0 {
		return
	}

	test()

}

func test() {
	c := bytecode.NewChunk()

	c.Write(123, bytecode.OP_CONSTANT, c.AddConstant(500))
	c.Write(123, bytecode.OP_NEGATE)
	c.Write(123, bytecode.OP_CONSTANT, c.AddConstant(600))
	c.Write(123, bytecode.OP_ADD)
	c.Write(123, bytecode.OP_RETURN)
	c.Disassemble()

	vm := vm.NewVM()
	result := vm.Interpret(c)

	log.Info("VM interpret finished", log.A("result", result))
}
