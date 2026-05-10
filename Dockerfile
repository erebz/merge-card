# Stage 1: build the binary.
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o merge-card ./cmd/merge-card

# Stage 2: minimal runtime image.
FROM alpine:3.20

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/merge-card ./merge-card

ENTRYPOINT ["/app/merge-card"]
