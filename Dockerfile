# Stage 1: Build stage
FROM golang:1.22.3-alpine3.19 AS builder

# Set working directory for the build stage
WORKDIR /usr/src/app

# Copy the entire application and build it
COPY . .

RUN go mod download && go mod verify

RUN go build -v -o echo ./cmd/echo/main.go

# Stage 2: Final stage
FROM alpine:3.19

# Set working directory for the final stage
WORKDIR /usr/local/bin

# Copy the built application and the config archive from the builder stage
COPY --from=builder /usr/src/app/echo echo

ENTRYPOINT [ "echo" ]