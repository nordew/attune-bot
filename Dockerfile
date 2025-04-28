FROM golang:1.24.1-alpine AS builder

RUN apk update && apk add --no-cache git ca-certificates tzdata

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY ./migrations/ ./migrations/
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/main ./cmd/app/main.go


FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

RUN adduser -D -g '' appuser

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /src/migrations /app/migrations
COPY .env /app/.env

RUN chown -R appuser:appuser /app

USER appuser

ENTRYPOINT ["./main"]
