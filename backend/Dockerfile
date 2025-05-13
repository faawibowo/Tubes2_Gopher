FROM golang:1.23 AS builder

WORKDIR /backend-build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o api-server ./cmd/main.go

FROM alpine:latest

WORKDIR /usr/local/backend

RUN apk add --no-cache tzdata

COPY --from=builder /backend-build/api-server .

COPY ./configs/elements.json ./configs/elements.json

EXPOSE 8080

ENTRYPOINT ["./api-server"]
