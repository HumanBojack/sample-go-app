FROM golang:1.22.1 as builder

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /build/main .
COPY --from=builder /build/templates/ ./templates

CMD ["./main"]
