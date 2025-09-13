package main

import (
	"bufio"
	"fmt"
	"internal/ast"
	"internal/bytecode"
	"internal/codegen"
	interpreter_ "internal/interpreter"
	"internal/parser"
	"internal/scanner"
	"internal/util/log"
	vm_ "internal/vm"
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

	run(fileBody, nil, nil)
}

func runLoop() {
	inputScanner := bufio.NewScanner(os.Stdin)
	interpreter := interpreter_.NewInterpreter()
	vm := vm_.NewVM()

	log.StdOut("> ")
	for inputScanner.Scan() {
		line := inputScanner.Bytes()
		run(line, interpreter, vm)
		log.StdOut("> ")
	}

	if err := inputScanner.Err(); err != nil {
		log.Fatal("Scanner error", log.E(err))
	}
}

func run(source []byte, interpreter *interpreter_.Interpreter, vm *vm_.VM) {
	sourceStr := string(source)

	log.InfoIfEnabled("Run source", func() []log.Field {
		_sourceStr := sourceStr

		if len(source) > 100 {
			_sourceStr = string(source[:100]) + "...(more " + fmt.Sprint(len(source)) + " bytes)"
		}

		return []log.Field{log.S("source", _sourceStr)}
	})

	// ================================================================
	// Scan
	// ================================================================
	scanner := scanner.NewScanner(sourceStr)
	tokens, errs := scanner.ScanTokens()

	log.Debug("Scan complete", log.A("tokens", tokens), log.A("errors", errs))

	if len(errs) > 0 {
		return
	}

	// ================================================================
	// Parse
	// ================================================================
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

	// ================================================================
	// Resolve + Interpret (HoLang1)
	// ================================================================
	if interpreter == nil {
		interpreter = interpreter_.NewInterpreter()
	}

	resolver := interpreter_.NewResolver(interpreter)
	err := resolver.Resolve(statements)

	log.Debug("Resolve complete", log.E(err))

	if err == nil {
		err = interpreter.Interpret(statements)

		log.Debug("Interpret complete", log.E(err))
	} else {
		log.Error("Resolve error", log.E(err))
	}

	// ================================================================
	// Codegen
	// ================================================================

	ch := bytecode.NewChunk()
	em := codegen.NewChunkEmitter(ch)
	gen := codegen.NewCodeGenerator(em)

	if err := gen.Generate(statements); err != nil {
		log.Error("Codegen error", log.E(err))

		return
	}

	disassemble := ch.Disassemble()
	log.Debug("Codegen complete", log.A("bytecode", disassemble))

	// ================================================================
	// Run
	// ================================================================
	if vm == nil {
		vm = vm_.NewVM()
	}
	result := vm.Interpret(ch)

	log.Info("VM interpret finished", log.A("result", result))

}
