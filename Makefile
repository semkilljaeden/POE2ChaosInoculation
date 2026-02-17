BINARY  = poe2crafter.exe
CMD     = ./cmd/poe2crafter

.PHONY: build run run-web clean tidy vet

build:
	go build -o $(BINARY) $(CMD)

run: build
	./$(BINARY)

run-web: build
	./$(BINARY) --web

run-debug: build
	./$(BINARY) --web --debug

clean:
	rm -f $(BINARY)

tidy:
	go mod tidy

vet:
	go vet ./...
