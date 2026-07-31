.PHONY: run test test-unit test-e2e lint vet bench fuzz clean coverage

run:
	go run cmd/api/main.go

test:
	go test -race -count=1 ./...

test-unit:
	go test -race ./test/unit/... -v

test-e2e:
	go test -race ./test/e2e/... -v

lint:
	go vet ./...

bench:
	go test -bench=. -benchmem ./... -run=^$$

fuzz:
	go test -fuzz=FuzzUrgencyScore -fuzztime=30s ./test/unit/

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean:
	rm -f coverage.out coverage.html
