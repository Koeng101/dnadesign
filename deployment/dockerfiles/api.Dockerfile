FROM golang:1.21 AS builder
WORKDIR /app

# Copy the go source
COPY . .

# Download Go modules
RUN go mod download

# Build the Go app
RUN GOOS=linux go build -o dnadesignapi api/main.go

# Final Stage
FROM scratch

# Copy the binary from the builder stage
COPY --from=builder /app/dnadesignapi .

# Command to run the executable
ENTRYPOINT ["./dnadesignapi"]

