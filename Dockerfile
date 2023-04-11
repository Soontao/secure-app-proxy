FROM golang:alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o main .

FROM alpine
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/main .
CMD ["/app/main"]
