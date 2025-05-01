# Use an Alpine-based Golang image for the build stage.
FROM golang:1.23-alpine as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Install C build dependencies for sqlite3 using apk
# Need gcc, musl-dev (for CGO), and sqlite-dev
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Build the Go app
# CGO_ENABLED=1 is needed for go-sqlite3
# -o lemc: output file name
# -ldflags="-w -s": link flags to reduce binary size
RUN go build -ldflags="-w -s" -o ./tmp/lemc ./main.go

# Start a new stage from a minimal base image
FROM alpine:latest

# Install SQLite runtime libraries
RUN apk add --no-cache sqlite-libs

# Set working directory (optional but good practice)
WORKDIR /app

# Copy only the pre-built binary file from the builder stage
# Assets and migrations are embedded in the binary
COPY --from=builder /app/tmp/lemc /app/lemc

# Command to run the executable
CMD ["/app/lemc"] 