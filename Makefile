BINARY := samir
CMD    := ./cmd/$(BINARY)
ARGS   := run /etc/printer

.PHONY: build run test clean

build:
	@go build -o $(BINARY) $(CMD)

run: build
	@sudo ./$(BINARY) $(ARGS)

test:
	@go test ./...

clean:
	@rm -f $(BINARY)
