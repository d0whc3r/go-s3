FROM golang:1.14 AS build_base

#RUN apk add --no-cache git make gcc

# Set the Current Working Directory inside the container
WORKDIR /app

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

# Build the Go app
RUN make platform-linux

# Start fresh from a smaller image
FROM bitnami/minideb

COPY --from=build_base /app/build/gos3-linux-64/gos3 /app/gos3

# Run the binary program produced by `go install`
ENTRYPOINT ["/app/gos3"]
