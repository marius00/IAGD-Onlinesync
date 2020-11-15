.PHONY: build clean deploy

test:
	go test ./... || exit

build:
	env GOOS=linux go build -ldflags="-s -w" -o bin/download endpoints/download/download.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/migrate endpoints/migrate/migrate.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/partitions endpoints/partitions/partitions.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/remove endpoints/remove/remove.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/upload endpoints/upload/upload.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/logincheck endpoints/session/logincheck/logincheck.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/monolith endpoints/monolith/monolith.go

clean:
	rm -rf ./bin

deploy: clean build
	sls deploy --verbose
