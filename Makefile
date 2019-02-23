BINARY := esclean

.PHONY: darwin
darwin:
	mkdir -p release
	GOOS=darwin GOARCH=amd64 go build -o release/$(BINARY)-darwin-amd64 cmd/esclean/esclean.go

.PHONY: linux
linux:
	mkdir -p release
	GOOS=linux GOARCH=amd64 go build -o release/$(BINARY)-linux cmd/esclean/esclean.go

.PHONY: clean
clean:
	rm -rf release/*

.PHONY: install	
install:
	GOOS=darwin GOARCH=amd64 go install cmd/esclean/esclean.go
