FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /parts-api cmd/api/main.go

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

COPY --from=builder /parts-api /usr/local/bin/parts-api

EXPOSE 8080

ENV PORT=8080
ENV DB_PATH=/data/parts.db

CMD ["parts-api"]
