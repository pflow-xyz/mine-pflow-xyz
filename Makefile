BINARY := mine-pflow-xyz

build:
	go build -o $(BINARY) .

run: build
	./$(BINARY)

test:
	go test ./...

clean:
	rm -f $(BINARY)

.PHONY: build run test clean
