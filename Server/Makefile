.PHONY: build clean deploy

test:
	env DATABASE_USER=test DATABASE_PASSWORD=test DATABASE_HOST=localhost DATABASE_NAME=test go test ./... || exit

build:
	env GOOS=linux go build -ldflags="-s -w" -o bin/delete endpoints/delete/delete.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/download endpoints/download/download.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/migrate endpoints/migrate/migrate.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/remove endpoints/remove/remove.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/auth endpoints/session/auth/auth.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/login endpoints/session/login/login.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/logincheck endpoints/session/logincheck/logincheck.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/upload endpoints/upload/upload.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/monolith endpoints/monolith/monolith.go

#lint:
#    golangci-lint run -E golint -E gosec -E interfacer -E unconvert -E dupl -E goconst -E gocyclo -E maligned -E depguard -E misspell -E unparam -E nakedret -E prealloc -E scopelint -E gochecknoinits

clean:
	rm -rf ./bin

deploy: clean build
	sls deploy --verbose


# TODO: GO Lint! didn't i set up linting...very recently?