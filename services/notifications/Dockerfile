FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/notification .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/notification .
EXPOSE 3000
CMD ["./notification"]
