FROM golang:latest AS builder

WORKDIR /app

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 go build -o main .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]
