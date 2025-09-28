TOOLS = \
	golang.org/x/tools/cmd/stringer@latest

tools:
	for t in $(TOOLS); do \
		go install $$t; \
	done

generate: tools
	go generate ./internal/bytecode

all: generate
	GOOS=linux GOARCH=amd64 go build -o dist/holang cmd/holang
	GOOS=windows GOARCH=amd64 go build -o dist/holang.exe cmd/holang
	GOOS=darwin GOARCH=amd64 go build -o dist/holang_darwin_amd64 cmd/holang
	GOOS=darwin GOARCH=arm64 go build -o dist/holang_darwin_arm64 cmd/holang

	GOOS=linux GOARCH=amd64 go build -o dist/holangc cmd/holangc
	GOOS=windows GOARCH=amd64 go build -o dist/holangc.exe cmd/holangc
	GOOS=darwin GOARCH=amd64 go build -o dist/holangc_darwin_amd64 cmd/holangc
	GOOS=darwin GOARCH=arm64 go build -o dist/holangc_darwin_arm64 cmd/holangc