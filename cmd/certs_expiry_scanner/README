# certs-expiry-scanner

A Go-based tool for scanning SSL/TLS certificate expiry dates across multiple endpoints, exporting metrics for Prometheus monitoring.

> **Note:** This tool is **not production-ready** and is intended for **learning and demonstration purposes only**.

## Features

- Concurrently checks SSL/TLS certificate expiry for configured endpoints.
- Exposes metrics via HTTP for Prometheus scraping.
- Configurable via JSON file (Kubernetes ConfigMap supported).


## Usage

### 1. Build and Push Docker Image

```bash
make build_and_push
```

### 2. Deploy to Kubernetes

```bash
make deploy
```

### 3. Configuration

- `timeout`: Specifies the maximum time (in seconds) to wait for a connection to each endpoint before timing out.
- `insecure_skip_verify`: If set to `true`, SSL certificate verification is skipped, allowing connections to endpoints with self-signed or invalid certificates.
- `check_interval`: Defines how often (in seconds) the scanner checks the endpoints for certificate expiry.
- `endpoints`: A list of endpoint addresses (in the format `host:port`) to be scanned for SSL certificate expiry.

Example config:

```json
{
    "timeout": 3,
    "insecure_skip_verify": true,
    "check_interval": 10,
    "endpoints": [
        "facebook.com:443",
        "google.com:443"
    ]
}
```

### 4. Metrics

Metrics are exposed at `/metrics` on port `3005`.
