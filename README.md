# Worker

A production-ready Go worker service for address matching. The worker searches for corporations and matches addresses using AI-powered comparison.

## Features

- 🔍 **Address Matching** - AI-powered address comparison using OpenAI
- ⚙️ **Configuration Management** - Environment-based configuration with `.env` support
- 📝 **Structured Logging** - Detailed progress logging for job processing
- 🐳 **Docker Support** - Multi-stage Dockerfile for optimized images
- ☸️ **Kubernetes/Helm** - Complete Helm chart for deployment
- 🔄 **CI/CD** - GitHub Actions workflows for testing, building, and releasing
- 🔧 **Makefile** - Common development tasks
- 📦 **Standard Structure** - Follows Go project layout best practices

## Project Structure

```
.
├── cmd/
│   └── worker/          # Main application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── matcher/         # Address matching using AI
│   ├── scraper/         # Web scraping for entity data
│   ├── workflows/       # Business logic workflows
│   └── worker.go        # Worker job processing
├── charts/
│   └── worker/          # Helm chart for Kubernetes
│       ├── templates/   # Kubernetes manifests
│       ├── values.yaml  # Default configuration values
│       └── Chart.yaml   # Chart metadata
├── .github/
│   └── workflows/       # GitHub Actions CI/CD workflows
├── Dockerfile           # Docker build file
├── Makefile             # Common tasks
├── go.mod               # Go module file
└── README.md            # This file
```

## Getting Started

### Prerequisites

- Go 1.23.3 or higher
- Make (optional, for using Makefile commands)
- OpenAI API key (for address matching)

### Installation

1. Clone this repository:

```bash
git clone <your-repo-url>
cd davidAI
```

2. Install dependencies:

```bash
make deps
# or
go mod download
```

3. Copy the environment example file:

```bash
cp env.example .env
```

4. Update `.env` with your configuration:

```bash
# Environment
ENV=development

# Logging
LOG_LEVEL=info

# Application Name
APP_NAME=worker

# Worker Configuration
NUM_WORKERS=1

# OpenAI Configuration (required for address matching)
OPENAI_API_KEY=your_openai_api_key_here
```

### Running the Application

#### Using Make

```bash
# Build and run
make run

# Or just build
make build
./bin/worker
```

#### Using Go directly

```bash
go run ./cmd/worker
```

#### Using Docker

```bash
# Build image
docker build -t worker:latest .

# Run container
docker run \
  -e OPENAI_API_KEY=your_key_here \
  worker:latest
```

## Configuration

Configuration is managed through environment variables. See `env.example` for available options:

- `ENV` - Environment: development, production (default: `development`)
- `LOG_LEVEL` - Log level: debug, info, warn, error (default: `info`)
- `APP_NAME` - Application name (default: `worker`)
- `NUM_WORKERS` - Number of worker goroutines (default: `1`)
- `OPENAI_API_KEY` - OpenAI API key for address matching (required)
- `ZYTE_API_KEY` - Zyte API key for web scraping (optional, has default)

## Development

### Building

```bash
# Build binary
make build

# The binary will be in bin/worker
```

### Testing

```bash
# Run tests
make test

# Run tests with coverage
make test-coverage
```

### Development Workflow

```bash
# Lint and build
make dev

# Run the server
make run
```

### Makefile Commands

- `make build` - Build the application binary
- `make run` - Build and start the worker server
- `make dev` - Lint code and build binary
- `make test` - Run tests
- `make test-coverage` - Run tests with coverage report
- `make clean` - Clean build artifacts
- `make fmt` - Format code
- `make lint` - Lint code (requires golangci-lint)
- `make deps` - Download and tidy dependencies
- `make docker-build` - Build Docker image
- `make help` - Show all available commands

## CI/CD

This project includes GitHub Actions workflows for continuous integration and deployment with branch-based environment configuration.

### Workflows

#### CI Workflow (`.github/workflows/ci.yml`)

Runs on every push and pull request to `main`, `master`, or `staging` branches:

- **Test Job**: Runs Go tests
- **Lint Job**: Runs code quality checks with golangci-lint
- **Build Job**: Builds the application binaries

#### Deploy Workflow (`.github/workflows/deploy.yml`)

Automatically deploys based on branch:

- **Staging Branch**: Deploys to staging environment when code is pushed to `staging` branch
- **Main/Master Branch**: Deploys to production environment when code is pushed to `main` or `master` branch
- **Manual Deployment**: Can be triggered manually via GitHub Actions UI with environment selection

The workflow:
1. Determines the target environment based on the branch
2. Builds and pushes Docker images with appropriate tags (`staging` or `latest`)
3. Deploys to Kubernetes using Helm with environment-specific values files

### Branch Strategy

- **`staging` branch**: Automatically deploys to staging environment
  - Uses `values/values-staging.yaml` for configuration
  - Image tag: `staging`
  - Namespace: `staging`
  
- **`main`/`master` branch**: Automatically deploys to production environment
  - Uses `values/values-prod.yaml` for configuration
  - Image tag: `latest`
  - Namespace: `production`

### Using the CI/CD

1. **Automatic deployments**: 
   - Push to `staging` branch → deploys to staging
   - Push to `main`/`master` branch → deploys to production

2. **Image location**: Images are pushed to `ghcr.io/<your-username>/<repo-name>/worker`

3. **Image tags**:
   - `staging` - Builds from staging branch
   - `latest` - Builds from main/master branch
   - `<branch-name>` - Builds from specific branches
   - `<sha>` - Builds tagged with commit SHA

4. **Manual deployment**: Use the "Run workflow" button in GitHub Actions to manually trigger deployments

### Configuration

To enable automatic deployments, you'll need to:

1. Configure Kubernetes secrets in GitHub repository settings:
   - Add `KUBECONFIG` secret with your Kubernetes cluster configuration
   - Or configure `KUBE_CONFIG_DATA` secret

2. Update the deploy workflow (`.github/workflows/deploy.yml`) with your actual Helm deployment commands (currently commented out)

### Pulling the Image

```bash
# Login to GHCR
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# Pull the image
docker pull ghcr.io/<your-username>/<repo-name>/worker:latest
docker pull ghcr.io/<your-username>/<repo-name>/worker:staging
```

## Docker

### Build Image

```bash
docker build -t worker:latest .
```

### Run Container

```bash
docker run \
  -e OPENAI_API_KEY=your_key_here \
  -e NUM_WORKERS=1 \
  worker:latest
```

### Using Docker Compose

Update `docker-compose.yml` with your configuration, then:

```bash
docker-compose up -d
```

## Kubernetes Deployment

This project includes a complete Helm chart for deploying to Kubernetes with support for staging and production environments.

### Prerequisites

- Kubernetes cluster
- Helm 3.x installed
- kubectl configured

### Environment-Specific Deployments

The project includes separate Helm values files for different environments in the `charts/worker/values/` folder:

- **Development**: `charts/worker/values/values-dev.yaml` - Minimal resources, debug logging, single replica
- **Staging**: `charts/worker/values/values-staging.yaml` - Lower resource limits, debug logging, staging namespace
- **Production**: `charts/worker/values/values-prod.yaml` - Higher resource limits, info logging, production namespace

### Deploy with Helm

1. Build and push your Docker image to a registry (or use GHCR):

```bash
docker build -t ghcr.io/<your-username>/<repo-name>-worker:latest .
docker push ghcr.io/<your-username>/<repo-name>-worker:latest
```

2. Update `charts/worker/values.yaml` with your image repository:

```yaml
image:
  repository: ghcr.io/<your-username>/<repo-name>-worker
  tag: latest
```

3. Install the chart:

```bash
# Install with default values
helm install worker ./charts/worker

# Install to development environment
helm install worker ./charts/worker \
  --namespace development \
  --create-namespace \
  -f charts/worker/values.yaml \
  -f charts/worker/values/values-dev.yaml

# Install to staging environment
helm install worker ./charts/worker \
  --namespace staging \
  --create-namespace \
  -f charts/worker/values.yaml \
  -f charts/worker/values/values-staging.yaml

# Install to production environment
helm install worker ./charts/worker \
  --namespace production \
  --create-namespace \
  -f charts/worker/values.yaml \
  -f charts/worker/values/values-prod.yaml
```

4. Upgrade an existing release:

```bash
# Upgrade development
helm upgrade worker ./charts/worker \
  --namespace development \
  -f charts/worker/values.yaml \
  -f charts/worker/values/values-dev.yaml

# Upgrade staging
helm upgrade worker ./charts/worker \
  --namespace staging \
  -f charts/worker/values.yaml \
  -f charts/worker/values/values-staging.yaml

# Upgrade production
helm upgrade worker ./charts/worker \
  --namespace production \
  -f charts/worker/values.yaml \
  -f charts/worker/values/values-prod.yaml
```

5. Uninstall:

```bash
helm uninstall worker --namespace development
helm uninstall worker --namespace staging
helm uninstall worker --namespace production
```

### Helm Chart Features

The Helm chart includes:

- **Deployment** - Configurable replicas, resources, and health checks
- **Service** - ClusterIP, NodePort, or LoadBalancer
- **Ingress** - Optional ingress configuration
- **ConfigMap** - For non-sensitive configuration
- **Secret** - For sensitive data (API keys)
- **ServiceAccount** - For RBAC
- **HorizontalPodAutoscaler** - Optional autoscaling
- **Health Probes** - Liveness, readiness, and startup probes

## How It Works

1. **Submit Job**: Jobs are submitted to the worker with corporation name and target address
2. **Search Entities**: Worker searches for active entities matching the corporation name
3. **Get Addresses**: For each entity, retrieves the registered address
4. **Compare Addresses**: Uses AI (OpenAI) to compare each entity address with the target address
5. **Return Result**: Returns the first match found, or empty response if no match

The worker processes jobs asynchronously and provides detailed logging throughout the process.

## Best Practices

1. **Error Handling**: Always handle errors explicitly
2. **Logging**: Use structured logging with context
3. **Configuration**: Use environment variables for configuration
4. **Testing**: Write tests for your handlers and business logic
5. **Documentation**: Document your API endpoints
6. **Security**: Never commit `.env` files or API keys
7. **CI/CD**: Use the provided workflows for automated testing and building

## License

MIT

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request
