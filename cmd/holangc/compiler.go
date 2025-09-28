package main

import (
	"fmt"
	"internal/ast"
	"internal/codegen"
	"internal/parser"
	"internal/resolver"
	"internal/runtime"
	"internal/scanner"
	"internal/util/log"
	"internal/vm"
)

func compile(source []byte) *runtime.ObjFunction {
	sourceStr := string(source)

	log.InfoIfEnabled("Source", func() []log.Field {
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
		return nil
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
		return nil
	}

	// ================================================================
	// Resolve
	// ================================================================
	r := resolver.NewResolver()
	bindings, errs := r.Resolve(program)

	log.Debug("Resolve complete", log.A("bindings", bindings), log.A("errors", errs))

	if len(errs) > 0 {
		return nil
	}

	// ================================================================
	// Codegen
	// ================================================================
	gen := codegen.NewCodeGenerator(bindings)

	rootFn, err := gen.Generate(program)
	if err != nil {
		log.Error("Codegen error", log.E(err))

		return nil
	}

	log.Debug("Codegen complete")

	rootFn.Disassemble()

	return rootFn
}

func run(fn *runtime.ObjFunction) error {
	machine := vm.NewVM()
	result := machine.Run(fn)

	log.Info("VM interpret finished", log.A("result", result))

	return nil
}

// func runLoop() {
// 	inputScanner := bufio.NewScanner(os.Stdin)
// 	interpreter := interpreter_.NewInterpreter()
// 	vm := vm_.NewVM()

// 	log.StdOut("> ")
// 	for inputScanner.Scan() {
// 		line := inputScanner.Bytes()
// 		run(line, interpreter, vm)
// 		log.StdOut("> ")
// 	}

// 	if err := inputScanner.Err(); err != nil {
// 		log.Fatal("Scanner error", log.E(err))
// 	}
// }
