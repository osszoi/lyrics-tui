BIN = lyrics-tui

build:
	go build -o $(BIN)

run:
	go run main.go

release:
ifndef v
	$(error "Please specify a version number. Example: make release v=1.1.0")
endif
	go build -ldflags "-X main.Version=v$(v)" -o $(BIN)
	chmod +x $(BIN)
	gh release create v$(v) $(BIN) --generate-notes --target master
