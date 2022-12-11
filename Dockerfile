# Use the official Golang image to create a build artifact.
# This is based on Debian and sets the GOPATH to /go.
# https://hub.docker.com/_/golang
FROM golang:1.18 as builder

# Copy local code to the container image.
WORKDIR /go/src/app
COPY main.go .
COPY config.json .

# Build the Go app.
RUN go build -o app

# Use the official Alpine image for a lean production container.
# https://hub.docker.com/_/alpine
# https://docs.docker.com/develop/develop-images/multistage-build/#use-multi-stage-builds
FROM alpine:3
RUN apk add --no-cache ca-certificates

# Copy the Go app to the production image from the builder stage.
COPY --from=builder /go/src/app/app /app

# Run the Go app.
CMD ["/app"]
