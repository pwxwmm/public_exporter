# Build stage
FROM golang:1.24 AS builder

# Set the working directory inside the container
WORKDIR /app

# Set Go module proxy for faster downloads
ENV GOPROXY="https://mirrors.aliyun.com/goproxy/"
ENV GO111MODULE=on
# Disable CGO for static compilation
ENV CGO_ENABLED=0

# Copy the source code into the container
COPY . .

# Install dependencies and build the Go binary
RUN go build -o public_exporter -v ./cmd/main.go && chmod +x public_exporter

# Use Alpine as the final runtime environment
FROM alpine:latest

# Install necessary dependencies: bash, python3, pip, and common Python modules
RUN sed -i 's|dl-cdn.alpinelinux.org|mirrors.aliyun.com|g' /etc/apk/repositories \
    && apk add --no-cache ca-certificates bash python3 py3-pip 

RUN pip install --no-cache-dir --break-system-packages \
    requests \
    PyYAML 

# 确保虚拟环境中的 Python 和 pip 可用
ENV PATH="/opt/venv/bin:$PATH"
# Create the log directory and ensure proper permissions
RUN mkdir -p /var/log/exporter && chmod 777 /var/log/exporter

# Set the working directory inside the container
WORKDIR /app

# Copy the compiled Go binary from the builder image
COPY --from=builder /app/public_exporter .

# Ensure the Go binary has execute permission
RUN chmod +x /app/public_exporter

# Expose the port for Prometheus metrics endpoint
EXPOSE 5535

# Set the entrypoint to run the exporter
ENTRYPOINT ["/app/public_exporter"]
