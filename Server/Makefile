.PHONY: build clean deploy

test:
	env DATABASE_USER=test DATABASE_PASSWORD=test DATABASE_HOST=localhost DATABASE_NAME=test go test ./... || exit

build:
	env GOOS=linux go build -ldflags="-s -w" -o bin/index endpoints/index/index.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/secured endpoints/secured/secured.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/session_login endpoints/session/login/login.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/session_logout endpoints/session/logout/logout.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/monolith endpoints/monolith/monolith.go

clean:
	rm -rf ./bin

deploy: clean build
	sls deploy --verbose
