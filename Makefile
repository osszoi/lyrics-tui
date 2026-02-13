build:
	go build -o lyrics-tui

run:
	go run main.go

release:
ifndef v
	$(error "Please specify a version number. Example: make release v=0.1.0")
endif
	make build
	gh release create v$(v) ./lyrics-tui --generate-notes --target master
	rm ./lyrics-tui

		
