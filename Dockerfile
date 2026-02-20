# Build Stage
FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o executor main.go

# Final Stage
FROM alpine:3.19

WORKDIR /app

# Install Terraform (and other tools if needed)
RUN apk add --no-cache git curl unzip bash ca-certificates

# Terraform will be installed dynamically
# Ensure cache directory exists and is writable
RUN mkdir -p /home/app/.terrakube/terraform-versions && \
  chmod -R 777 /home/app

COPY --from=builder /app/executor .

# Default to Online Mode
ENV EXECUTOR_MODE=ONLINE
ENV PORT=8090

EXPOSE 8090

CMD ["./executor"]
