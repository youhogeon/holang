package main

import (
	"bufio"
	"fmt"
	"internal/ast"
	_interpreter "internal/interpreter"
	"internal/parser"
	"internal/resolver"
	"internal/scanner"
	"internal/util/log"
	"os"
)

func runFile(fileName string) {
	fileBody, err := os.ReadFile(fileName)
	if err != nil {
		log.Fatal("Read file error", log.S("file", fileName), log.E(err))
	}

	run(fileBody, nil)
}

func runLoop() {
	inputScanner := bufio.NewScanner(os.Stdin)
	interpreter := _interpreter.NewInterpreter()

	log.StdOut("> ")
	for inputScanner.Scan() {
		line := inputScanner.Bytes()
		run(line, interpreter)
		log.StdOut("> ")
	}

	if err := inputScanner.Err(); err != nil {
		log.Fatal("Scanner error", log.E(err))
	}
}

func run(source []byte, interpreter *_interpreter.Interpreter) {
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

	statements, errs := p.Parse()

	log.Debug("Parse complete", log.A("ast", statements), log.A("errors", errs))

	for _, stmt := range statements {
		log.Debug("AST", log.S("astStr", printer.PrintStmt(stmt)))
	}

	if interpreter == nil {
		interpreter = _interpreter.NewInterpreter()
	}

	resolver := resolver.NewResolver(interpreter)
	err := resolver.Resolve(statements)

	log.Debug("Resolve complete", log.E(err))

	if err != nil {
		return
	}

	err = interpreter.Interpret(statements)

	log.Debug("Interpret complete", log.E(err))
}
