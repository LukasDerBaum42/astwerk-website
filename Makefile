templ=$(HOME)/go/bin/templ

main:
	$(templ) generate
	go run .

dev:
	$(templ) generate
	go run . --serve

test:
	go test ./...