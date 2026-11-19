# Build Stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o cv-generator ./cmd/api

# Final Stage
FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/cv-generator /app/cv-generator
COPY configs/ /app/configs/
COPY templates/ /app/templates/

EXPOSE 8080

CMD ["/app/cv-generator"]

