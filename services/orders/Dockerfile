FROM golang:1.25.4-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o orders-service ./main.go

EXPOSE 3003

CMD ["./orders-service"]