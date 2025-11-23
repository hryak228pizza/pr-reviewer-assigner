FROM golang:1.25-alpine AS builder

RUN apk update && apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/pr-reviewer-assigner ./cmd/main.go

FROM alpine:3.19

RUN apk update && apk add --no-cache tzdata

WORKDIR /app

COPY --from=builder /app/pr-reviewer-assigner /app/pr-reviewer-assigner

ENTRYPOINT ["/app/pr-reviewer-assigner"]