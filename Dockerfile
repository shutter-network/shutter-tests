# Start from a Debian based image with Go installed
FROM golang:1.23

# Set the current working directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all the dependencies
RUN go mod download

# Copy the source from the current directory to the working directory inside the container
COPY config /app/config
COPY continuous /app/continuous
COPY requests /app/requests
COPY stress /app/stress
COPY tests /app/tests
COPY utils /app/utils
COPY main.go /app/

VOLUME /app/logs
# Build the Go app
RUN go build -o main .

# Run the executable
ENTRYPOINT ["/app/main"]
