FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /notifications ./cmd/api

FROM alpine:3.23

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /notifications .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/resources ./resources

EXPOSE 8081

CMD ["./notifications"]
