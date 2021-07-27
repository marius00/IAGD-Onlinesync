# Start by building the application.
FROM golang:1.13-buster as build

WORKDIR /go/src/app
ADD . /go/src/app

RUN go get -d -v ./...

RUN env GOOS=linux go build -ldflags="-s -w" -o /go/bin/app endpoints/monolith.go

# Now copy it into our base image.
FROM gcr.io/distroless/base-debian10
COPY --from=build /go/bin/app /
EXPOSE 8080
ENV DATABASE_HOST=
ENV DATABASE_NAME=
ENV DATABASE_USER=
ENV DATABASE_PASSWORD=
ENV ALLOWED_ORIGIN=
ENV GO_API_HOST=
ENV REGION=
ENV BUCKETNAME=
CMD ["/app"]