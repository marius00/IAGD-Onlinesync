.PHONY: build clean deploy

test:
	env SECRET=test go test ./... || exit

build:
	env GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o bin/sendmail endpoints/sendmail/sendmail.go
	env GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o bin/monolith endpoints/monolith/monolith.go

clean:
	rm -rf ./bin

install: clean build

deploy: clean build
	sls deploy --verbose
