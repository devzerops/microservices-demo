FROM golang:1.23.3-alpine@sha256:c694a4d291a13a9f9d94933395673494fc2cc9d4777b85df3a7e70b3492d3574 AS builder

WORKDIR /src

# Restore dependencies
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the service
RUN go build -o /authservice .

FROM alpine:3.21@sha256:21dc6063fd678b478f57c0e13f47560d0ea4eeba26dfc947b2a4f81f686b9f45 AS release

RUN apk add --no-cache ca-certificates

WORKDIR /src

COPY --from=builder /authservice /src/authservice

# Enable health checks
RUN apk add --no-cache curl

EXPOSE 8081
ENTRYPOINT ["/src/authservice"]
