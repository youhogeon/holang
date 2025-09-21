package main

import (
	"bufio"
	"bytes"
	"fmt"
	"internal/ast"
	"internal/codegen"
	interpreter_ "internal/interpreter"
	"internal/parser"
	"internal/resolver"
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

	program, errs := p.Parse()
	statements := program.Statements

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

	r_ := interpreter_.NewResolver(interpreter)
	err := r_.Resolve(program)

	log.Debug("Resolve complete", log.E(err))

	if err == nil {
		err = interpreter.Interpret(program)

		log.Debug("Interpret complete", log.E(err))
	}

	// ================================================================
	// Resolve
	// ================================================================
	r := resolver.NewResolver()
	bindings, errs := r.Resolve(program)

	log.Debug("Resolve complete", log.A("bindings", bindings), log.A("errors", errs))

	if len(errs) > 0 {
		return
	}

	// ================================================================
	// Codegen
	// ================================================================
	gen := codegen.NewCodeGenerator(bindings)

	rootFn, err := gen.Generate(program)
	if err != nil {
		log.Error("Codegen error", log.E(err))

		return
	}

	log.Debug("Codegen complete")

	rootFn.Disassemble()

	// ================================================================
	// Run
	// ================================================================
	if vm == nil {
		vm = vm_.NewVM()
	}
	result := vm.Run(rootFn)

	log.Info("VM interpret finished", log.A("result", result))

	// ================================================================
	// Serialize
	// ================================================================
	var bytecode bytes.Buffer

	if err := rootFn.Serialize(&bytecode); err != nil {
		log.Error("Bytecode serialize error", log.E(err))

		return
	}

	log.Debug("Bytecode serialize complete", log.I("size", bytecode.Len()), log.A("bytecode", bytecode.Bytes()))
}
