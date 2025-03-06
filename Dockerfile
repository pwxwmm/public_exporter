FROM golang:1.20 AS builder

WORKDIR /app
COPY . .

RUN go mod tidy && go build -o publice_exporter cmd/main.go

FROM ubuntu:20.04
WORKDIR /app
COPY --from=builder /app/publice_exporter .

EXPOSE 5535
ENTRYPOINT ["/app/publice_exporter"]