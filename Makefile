all:
	GOOS=linux GOARCH=amd64 go build -o dist/holang_linux_amd64 cmd/main
	GOOS=windows GOARCH=amd64 go build -o dist/holang.exe cmd/main
	GOOS=darwin GOARCH=amd64 go build -o dist/holang_darwin_amd64 cmd/main
	GOOS=darwin GOARCH=arm64 go build -o dist/holang_darwin_arm64 cmd/main