.PHONY: clean, test, test-coverage, all

build:
	go build -o ./bin/app.exe ./cmd/
clean:
	rm -rf ./bin/app.exe
	rm -rf ./coverage ./rbac_verify.log
test:
	go test ./...
test-coverage:
	go test ./... -coverprofile=coverage.out
all:
	make clean
	make build
	make test
