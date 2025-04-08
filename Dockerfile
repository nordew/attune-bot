FROM golang:1.24.1-alpine AS builder

RUN apk update && apk add --no-cache git

WORKDIR /src

COPY ./migrations/ ./migrations/
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./cmd/app/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/main .

EXPOSE 8080

ENTRYPOINT ["./main"]
