# Simple Service Account Token Issuer

A simple HTTP server for creating Kubernetes service account tokens. This service allows you to issue temporary tokens for different service accounts with predefined roles (read-only or admin) through a REST API.

## Overview

This application provides a secure way to issue Kubernetes service account tokens programmatically. It runs as a service within your Kubernetes cluster and can issue tokens with different permission levels based on the requested role.

## Features

- Issues temporary Kubernetes service account tokens
- Supports different roles (read-only and admin)
- Configurable token expiration time
- Authentication protection for token issuance
- Kubernetes-native deployment

## Prerequisites

- Go 1.24 or later (for development)
- Docker (for building container images)
- Kubernetes cluster (for deployment)
- kubectl configured to access your cluster

## Local Development

### Setup

1. Clone the repository:

```bash
git clone https://github.com/zezaeoh/simple-sa-token-issuer.git
cd simple-sa-token-issuer
```

2. Build the application:

```bash
make build
```

3. Run locally (Note: This will not work fully outside a Kubernetes cluster):

```bash
make run
```

### Docker Development

For local testing with Docker:

```bash
# Build the Docker image
make docker-build

# Run with Docker
make docker-run
```

Alternatively, use Docker Compose:

```bash
# Start the service
make compose-up

# Stop the service
make compose-down
```

## Configuration

The application can be configured using environment variables:

| Variable | Description | Default |
|----------|-------------|--------|
| PORT | HTTP server port | 8080 |
| AUTH_TOKEN | Authentication token for API requests | "" (empty, no auth) |
| READONLY_SA | Service account name for read-only role | "readonly" |
| READONLY_NS | Namespace of read-only service account | "default" |
| ADMIN_SA | Service account name for admin role | "admin" |
| ADMIN_NS | Namespace of admin service account | "default" |

## Deployment to Kubernetes

### 1. Create the necessary resources

Apply the RBAC configuration to create service accounts and roles:

```bash
kubectl apply -f k8s/rbac.yaml
```

This creates:
- A `token-issuer-sa` service account with permission to create tokens
- `readonly` and `admin` service accounts with appropriate permissions
- Necessary ClusterRoles and ClusterRoleBindings

### 2. Create the authentication secret

Apply the secret configuration (or create your own with a secure token):

```bash
kubectl apply -f k8s/secret.yaml
```

Or create a custom secret:

```bash
kubectl create secret generic token-issuer-secret \
  --from-literal=auth-token=your-secure-token-here
```

### 3. Deploy the application

```bash
kubectl apply -f k8s/deployment.yaml
```

## API Usage

### Health Check

```bash
curl http://<service-address>:8080/healthz
```

### Request a Token

For a read-only token:

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Token your-auth-token" \
  -d '{"role": "read-only"}' \
  http://<service-address>:8080/token
```

For an admin token:

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Token your-auth-token" \
  -d '{"role": "admin"}' \
  http://<service-address>:8080/token
```

Response format:

```json
{
  "kind": "ExecCredential",
  "apiVersion": "client.authentication.k8s.io/v1beta1",
  "token": "<kubernetes-service-account-token>",
  "expirationTimestamp": "2023-01-01T00:00:00Z"
}
```

## Security Considerations

- Always use HTTPS in production
- Set a strong AUTH_TOKEN value
- Deploy in a secure namespace with appropriate network policies
- Consider using a service mesh for additional security layers

## License

See the [LICENSE](LICENSE) file for details.
