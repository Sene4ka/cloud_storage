# cloud_storage
Simple cloud storage backend realisation

**Stack**:
- Go
- gRPC for microservice communication
- PgSQL
- Redis
- MinIO

**Deploy**: 
- Docker + Docker Compose

**Observability**:
- Prometheus
- Graphana

## **Starting API on Your Machine Tutorial**

```bash
# 1. Clone + setup
git clone https://github.com/Sene4ka/cloud_storage.git
cd cloud_storage

# 2. Environment
cp .env.example .env
# Set your desired env values for docker compose

# 3. Start the server
make docker-up # Automatically Executes 'make proto' and 'make build' before starting API with 'docker compose'

# 4. For more options info
make help```
