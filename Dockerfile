FROM golang:1.22 AS build_base

#RUN apk add --no-cache git make gcc

# Set the Current Working Directory inside the container
WORKDIR /app

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

# Build the Go app
RUN go build "-ldflags=-s -w" -o /app/build-app

# Start fresh from a smaller image
FROM bitnami/minideb

COPY --from=build_base /app/build-app /app/gos3

# Run the binary program produced by `go install`
ENTRYPOINT ["/app/gos3"]
