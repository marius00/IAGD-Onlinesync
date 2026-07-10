.PHONY: build clean docker test

# Requires a local MySQL mirror (see docker-compose in D:\Dev\item).
test:
	env DATABASE_USER=root DATABASE_PASSWORD=123456 DATABASE_HOST=localhost DATABASE_NAME=iagd_backup go test ./... || exit

build:
	rm -f bin/*
	env GOOS=linux go build -ldflags="-s -w" -o bin/monolith endpoints/monolith.go

# Build the Docker image deployed to Coolify.
docker:
	docker build -t iagdbackup .

#lint:
#    golangci-lint run -E golint -E gosec -E interfacer -E unconvert -E dupl -E goconst -E gocyclo -E maligned -E depguard -E misspell -E unparam -E nakedret -E prealloc -E scopelint -E gochecknoinits

clean:
	rm -rf ./bin