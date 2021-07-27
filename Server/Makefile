.PHONY: build clean deploy

test:
	env DATABASE_USER=test DATABASE_PASSWORD=test DATABASE_HOST=localhost DATABASE_NAME=test go test ./... || exit

build:
	rm -f bin/*
	env GOOS=linux go build -ldflags="-s -w" -o bin/monolith endpoints/monolith.go

#lint:
#    golangci-lint run -E golint -E gosec -E interfacer -E unconvert -E dupl -E goconst -E gocyclo -E maligned -E depguard -E misspell -E unparam -E nakedret -E prealloc -E scopelint -E gochecknoinits

clean:
	rm -rf ./bin

deploy: clean build
	sls deploy --verbose


# TODO: GO Lint! didn't i set up linting...very recently?