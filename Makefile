build:
	go mod tidy
	go build -o bin/flight-info-agg ./cmd/app

run: build
	./bin/flight-info-agg

clean:
	go clean
	rm bin/flight-info-agg

dev:
	go run main.go


