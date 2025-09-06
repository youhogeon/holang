package main

import (
	"bufio"
	"fmt"
	"internal/ast"
	"internal/parser"
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

	lex := scanner.NewScanner(sourceStr)
	tokens, errs := lex.ScanTokens()

	log.Debug("Scan complete", log.A("tokens", tokens), log.A("errors", errs))

	if len(errs) > 0 {
		return
	}

	p := parser.NewParser(tokens)
	printer := ast.NewAstPrinter()

	root, err := p.Parse()

	log.Debug("Parse complete", log.A("ast", root), log.E(err))

	log.Debug("AST", log.S("astStr", printer.PrintExpr(root)))
}
