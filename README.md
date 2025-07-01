# Simple Service Account Token Issuer

A simple HTTP server for creating Kubernetes service account tokens. This service allows you to issue temporary tokens for different service accounts with configurable roles through a REST API.

## Overview

This application provides a secure way to issue Kubernetes service account tokens programmatically. It runs as a service within your Kubernetes cluster and can issue tokens with different permission levels based on dynamically configured roles.

## Features

- Issues temporary Kubernetes service account tokens
- Supports dynamically configurable roles via JSON configuration
- Configurable token expiration time
- Authentication protection for token issuance
- Kubernetes-native deployment
- Extensible role-based access control

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
| ROLES_CONFIG | JSON configuration for available roles | Default roles config |

### Role Configuration (ROLES_CONFIG)

The `ROLES_CONFIG` environment variable should contain a JSON object mapping role names to their corresponding service account configurations:

```json
{
  "read-only": {
    "serviceAccount": "readonly",
    "namespace": "default"
  },
  "admin": {
    "serviceAccount": "admin", 
    "namespace": "default"
  },
  "developer": {
    "serviceAccount": "developer",
    "namespace": "development"
  },
  "monitoring": {
    "serviceAccount": "monitoring",
    "namespace": "monitoring"
  }
}
```

Each role configuration contains:
- `serviceAccount`: The name of the Kubernetes service account to use for this role
- `namespace`: The namespace where the service account is located

## Deployment to Kubernetes

### 1. Create the necessary resources

Apply the RBAC configuration to create service accounts and roles:

```bash
kubectl apply -f k8s/rbac.yaml
```

This creates:
- A `token-issuer-sa` service account with permission to create tokens
- Example service accounts (`readonly`, `admin`, `developer`, `monitoring`) with appropriate permissions
- Necessary ClusterRoles, Roles, and bindings
- Example namespaces (`development`, `monitoring`) for demonstration

### 2. Create the role configuration

Apply the ConfigMap that contains the role configuration:

```bash
kubectl apply -f k8s/configmap.yaml
```

This creates a ConfigMap with the default role configuration. You can modify this ConfigMap to add, remove, or update roles as needed.

### 3. Create the authentication secret

Apply the secret configuration (or create your own with a secure token):

```bash
kubectl apply -f k8s/secret.yaml
```

Or create a custom secret:

```bash
kubectl create secret generic token-issuer-secret \
  --from-literal=auth-token=your-secure-token-here
```

### 4. Deploy the application

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

For a developer token:

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Token your-auth-token" \
  -d '{"role": "developer"}' \
  http://<service-address>:8080/token
```

For a monitoring token:

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Token your-auth-token" \
  -d '{"role": "monitoring"}' \
  http://<service-address>:8080/token
```

Response format:

```json
{
  "kind": "ExecCredential",
  "apiVersion": "client.authentication.k8s.io/v1beta1",
  "spec": {
    "interactive": false
  },
  "status": {
    "token": "<kubernetes-service-account-token>",
    "expirationTimestamp": "2023-01-01T00:00:00Z"
  }
}
```

## Managing Roles

### Adding New Roles

To add a new role:

1. Create the necessary service account and RBAC resources in your cluster
2. Update the ConfigMap to include the new role:

```bash
kubectl patch configmap token-issuer-config --patch '
{
  "data": {
    "roles-config": "{\"read-only\":{\"serviceAccount\":\"readonly\",\"namespace\":\"default\"},\"admin\":{\"serviceAccount\":\"admin\",\"namespace\":\"default\"},\"developer\":{\"serviceAccount\":\"developer\",\"namespace\":\"development\"},\"monitoring\":{\"serviceAccount\":\"monitoring\",\"namespace\":\"monitoring\"},\"new-role\":{\"serviceAccount\":\"new-sa\",\"namespace\":\"new-namespace\"}}"
  }
}'
```

3. Restart the deployment to pick up the configuration changes:

```bash
kubectl rollout restart deployment simple-sa-token-issuer
```

### Removing Roles

To remove a role, simply update the ConfigMap to exclude that role and restart the deployment.

## Security Considerations

- Always use HTTPS in production
- Set a strong AUTH_TOKEN value
- Deploy in a secure namespace with appropriate network policies
- Consider using a service mesh for additional security layers
- Regularly review and audit the configured roles and their permissions
- Follow the principle of least privilege when creating service account permissions

## Examples

The default configuration includes these example roles:

- **read-only**: Basic read access to common Kubernetes resources
- **admin**: Full administrative access to the cluster
- **developer**: Full access within the `development` namespace only
- **monitoring**: Read-only access to resources needed for monitoring

These examples demonstrate the flexibility of the role-based configuration system.

## License

See the [LICENSE](LICENSE) file for details.
