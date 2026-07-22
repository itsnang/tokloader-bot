FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -ldflags="-w -s" -o out

FROM alpine:3.21

WORKDIR /app

COPY --from=builder /app/out .

CMD ["./out"]
