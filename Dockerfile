# syntax=docker/dockerfile:1

FROM golang:1.23 as builder

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./go-gcsproxy

FROM alpine:latest

COPY --from=builder /app/go-gcsproxy /

# Optional:
# To bind to a TCP port, runtime parameters must be supplied to the docker command.
# But we can document in the Dockerfile what ports
# the application is going to listen on by default.
# https://docs.docker.com/reference/dockerfile/#expose
EXPOSE 9080

# Run
CMD ["/go-gcsproxy"]