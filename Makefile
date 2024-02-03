build:
	go mod tidy
	go build -o bin/backend ./cmd/app
	go build -o bin/cron ./cmd/cron

run: build
	./bin/backend

run-cron: build
	./bin/cron

clean:
	go clean
	rm bin/flight-info-agg

dev:
	go run cmd/app/main.go

dev-cron:
	go run cmd/cron/main.go

