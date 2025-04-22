FROM golang:1.24-alpine3.21 AS builder

RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main ./cmd/main.go

FROM alpine:3.21

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /go/bin/migrate /app/migrate
COPY db/migration ./db/migration
COPY app.env .

RUN chmod +x /app/main /app/migrate

EXPOSE 8080

CMD sh -c "/app/migrate -path db/migration -database \"postgresql://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable\" -verbose up && ./main"
