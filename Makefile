BIN = lyrics-tui

build:
	go build -o $(BIN)

run:
	go run main.go

release:
ifndef v
	$(error "Please specify a version number. Example: make release v=0.1.0")
endif
	make build
	chmod +x $(BIN)
	gh release create v$(v) $(BIN) --generate-notes --target master

