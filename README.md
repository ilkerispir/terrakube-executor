# Terrakube Executor (Go)

This is the Go implementation of the [Terrakube Executor](https://terrakube.org/). It is designed to be a lightweight, high-performance replacement for the original Java-based executor.

## Features

*   **High Performance**: Written in Go for low memory footprint and fast startup times.
*   **Dual Execution Modes**:
    *   **ONLINE**: Long-running HTTP server that receives jobs via API.
    *   **BATCH**: Ephemeral execution mode for Kubernetes Jobs (reads job data from env var).
*   **Dynamic Terraform/OpenTofu Management**:
    *   Automatically downloads the required Terraform/OpenTofu version for each job using `hashicorp/hc-install`.
    *   Supports both AMD64 (Linux) and ARM64 (Apple Silicon) architectures.
    *   Caches binaries locally to avoid repeated downloads.
*   **Workspace Management**:
    *   Clones Git repositories (public & private).
    *   Handles SSH keys and Access Tokens.
*   **Storage Backend**:
    *   Supports AWS S3, Azure Blob Storage, and Google Cloud Storage (GCS) for state and plan files.
*   **Real-time Logging**: Streams logs to **Redis Streams** for live UI updates.

## Configuration

The executor is configured via environment variables:

| Variable | Description | Default |
| :--- | :--- | :--- |
| `EXECUTOR_MODE` | Execution mode: `ONLINE` or `BATCH` | `ONLINE` |
| `PORT` | HTTP Port for `ONLINE` mode | `8080` |
| `TERRAKUBE_API_URL` | URL of the Terrakube API | (Required) |
| `STORAGE_TYPE` | Storage backend: `AWS`, `AZURE`, `GCP` | (Required) |
| `EPHEMERAL_JOB_DATA` | Base64 encoded JSON job data (for `BATCH` mode) | (Required for Batch) |

### Storage Configuration

**AWS S3:**
*   `AWS_REGION`
*   `AWS_ACCESS_KEY_ID`
*   `AWS_SECRET_ACCESS_KEY`
*   `AWS_BUCKET_NAME`

**Azure Blob Storage:**
*   `AZURE_STORAGE_ACCOUNT_NAME`
*   `AZURE_STORAGE_ACCOUNT_KEY`
*   `AZURE_STORAGE_CONTAINER_NAME`

**Google Cloud Storage:**
*   `GCP_STORAGE_BUCKET`
*   `GCP_SERVICE_ACCOUNT_KEY` (Path to JSON key file or content)

### Redis Configuration (Logs)
*   `USE_REDIS_LOGS`: `true` or `false`
*   `REDIS_HOST`: Redis host address
*   `REDIS_PASSWORD`: Redis password

## Local Development

### Prerequisites
*   Go 1.22+
*   Docker (optional)

### Build
```bash
go mod download
go build -o executor main.go
```

### Run (Online Mode)
```bash
export EXECUTOR_MODE=ONLINE
export PORT=8080
export TERRAKUBE_API_URL="http://localhost:8080"
export STORAGE_TYPE="AWS" # Example
./executor
```

### Run (Batch Mode)
```bash
export EXECUTOR_MODE=BATCH
export EPHEMERAL_JOB_DATA="<base64_encoded_job_json>"
./executor
```

## Docker

Build the Docker image:

```bash
docker build -t terrakube-executor-go .
```

Run container:

```bash
docker run -p 8080:8080 -e EXECUTOR_MODE=ONLINE terrakube-executor-go
```
