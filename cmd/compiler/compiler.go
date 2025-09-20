package main

import (
	"bytes"
	"fmt"
	"internal/ast"
	"internal/codegen"
	"internal/parser"
	"internal/resolver"
	"internal/scanner"
	"internal/util/log"
)

func compile(source []byte) []byte {
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

	var bytecode bytes.Buffer

	if err := rootFn.Serialize(&bytecode); err != nil {
		log.Error("Bytecode serialize error", log.E(err))

		return nil
	}

	log.Debug("Bytecode serialize complete", log.I("size", bytecode.Len()), log.A("bytecode", bytecode.Bytes()))

	return bytecode.Bytes()
}
