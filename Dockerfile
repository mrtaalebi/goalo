FROM golang:latest as BUILD

WORKDIR /app

COPY go.mod .
COPY go.sum .
COPY vendor ./vendor
COPY cmd ./cmd
COPY pkg ./pkg

RUN CGO_ENABLED=1 GOOS=linux go build -o goalo cmd/goalo/main.go

FROM ubuntu:20.04

WORKDIR /app
COPY --from=BUILD /app/goalo .

EXPOSE 80

ENTRYPOINT ["./goalo"]
