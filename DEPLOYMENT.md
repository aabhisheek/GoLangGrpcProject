# Deployment Guide

This guide covers deploying the Voucher & Payment Service to various environments.

## Table of Contents

- [Local Development](#local-development)
- [Docker Deployment](#docker-deployment)
- [Kubernetes Deployment](#kubernetes-deployment)
- [Production Checklist](#production-checklist)
- [Monitoring](#monitoring)

## Local Development

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- Make

### Setup

```bash
# Start infrastructure
make docker-up

# Run migrations
make migrate-up

# Generate proto files
make proto

# Run the service
make run
```

## Docker Deployment

### Build Image

```bash
docker build -t voucher-payment-service:v1.0.0 .
```

### Run with Docker Compose

```bash
docker-compose up -d
```

### Environment Variables

Create a `.env` file:

```env
# Database
DB_HOST=mysql
DB_PORT=3306
DB_USER=voucheruser
DB_PASSWORD=voucherpass
DB_NAME=voucher_db
DB_MAX_CONNECTIONS=25

# Redis
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_POOL_SIZE=10

# RabbitMQ
RABBITMQ_HOST=rabbitmq
RABBITMQ_PORT=5672
RABBITMQ_USER=guest
RABBITMQ_PASSWORD=guest
RABBITMQ_EXCHANGE=voucher.events
RABBITMQ_QUEUE=voucher.purchases

# Observability
JAEGER_ENDPOINT=http://jaeger:14268/api/traces
LOG_LEVEL=info
LOG_FORMAT=json

# Server
GRPC_PORT=50051
HTTP_PORT=8080
```

## Kubernetes Deployment

### Prerequisites

- Kubernetes cluster (1.24+)
- kubectl configured
- Helm 3

### Deploy with Helm

Create `values.yaml`:

```yaml
replicaCount: 3

image:
  repository: voucher-payment-service
  tag: v1.0.0
  pullPolicy: IfNotPresent

service:
  type: LoadBalancer
  grpcPort: 50051
  httpPort: 8080

resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 250m
    memory: 256Mi

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70

mysql:
  host: mysql-service
  port: 3306
  database: voucher_db

redis:
  host: redis-service
  port: 6379

rabbitmq:
  host: rabbitmq-service
  port: 5672
```

Deploy:

```bash
helm install voucher-service ./helm-chart -f values.yaml
```

### Kubernetes Manifests

**Deployment:**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: voucher-service
  labels:
    app: voucher-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: voucher-service
  template:
    metadata:
      labels:
        app: voucher-service
    spec:
      containers:
      - name: voucher-service
        image: voucher-payment-service:v1.0.0
        ports:
        - containerPort: 50051
          name: grpc
        - containerPort: 8080
          name: http
        env:
        - name: DB_HOST
          valueFrom:
            configMapKeyRef:
              name: voucher-config
              key: db-host
        resources:
          limits:
            cpu: 500m
            memory: 512Mi
          requests:
            cpu: 250m
            memory: 256Mi
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 5
```

**Service:**

```yaml
apiVersion: v1
kind: Service
metadata:
  name: voucher-service
spec:
  type: LoadBalancer
  ports:
  - port: 50051
    targetPort: 50051
    name: grpc
  - port: 8080
    targetPort: 8080
    name: http
  selector:
    app: voucher-service
```

**HorizontalPodAutoscaler:**

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: voucher-service-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: voucher-service
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

## Production Checklist

### Security

- [ ] Enable TLS/SSL for all services
- [ ] Implement proper authentication (JWT)
- [ ] Use secrets management (Vault, AWS Secrets Manager)
- [ ] Enable network policies
- [ ] Regular security audits
- [ ] Input validation and sanitization
- [ ] Rate limiting enabled
- [ ] SQL injection prevention

### Reliability

- [ ] Set up health checks
- [ ] Configure graceful shutdown
- [ ] Implement circuit breakers
- [ ] Set timeouts on all operations
- [ ] Database connection pooling
- [ ] Redis failover configuration
- [ ] RabbitMQ clustering
- [ ] Backup and recovery procedures

### Performance

- [ ] Enable caching layers
- [ ] Database query optimization
- [ ] Connection pooling configured
- [ ] Worker pool tuning
- [ ] Load testing completed
- [ ] CDN for static assets
- [ ] Compression enabled

### Observability

- [ ] Distributed tracing enabled
- [ ] Metrics collection (Prometheus)
- [ ] Log aggregation (ELK/Loki)
- [ ] Alerting configured
- [ ] Dashboard creation (Grafana)
- [ ] Error tracking (Sentry)

### Infrastructure

- [ ] Auto-scaling configured
- [ ] Multi-AZ deployment
- [ ] Disaster recovery plan
- [ ] Database replication
- [ ] Regular backups
- [ ] Infrastructure as Code (Terraform)

## Monitoring

### Prometheus Metrics

Exposed at `/metrics`:

- `http_requests_total` - Total HTTP requests
- `http_request_duration_seconds` - Request duration
- `grpc_requests_total` - Total gRPC requests
- `db_connections_active` - Active DB connections
- `cache_hits_total` - Cache hit count
- `cache_misses_total` - Cache miss count

### Grafana Dashboards

Import dashboard IDs:
- Go Application Metrics: 6671
- MySQL Overview: 7362
- Redis Overview: 11835
- RabbitMQ Overview: 4279

### Alerts

Example Prometheus alerts:

```yaml
groups:
- name: voucher-service
  rules:
  - alert: HighErrorRate
    expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.05
    for: 5m
    annotations:
      summary: "High error rate detected"
      
  - alert: HighLatency
    expr: histogram_quantile(0.95, http_request_duration_seconds) > 1
    for: 5m
    annotations:
      summary: "High latency detected"
      
  - alert: LowCacheHitRate
    expr: rate(cache_hits_total[5m]) / (rate(cache_hits_total[5m]) + rate(cache_misses_total[5m])) < 0.8
    for: 10m
    annotations:
      summary: "Cache hit rate below 80%"
```

## Scaling Strategies

### Horizontal Scaling

- Add more service replicas
- Use load balancer
- Stateless design

### Vertical Scaling

- Increase CPU/memory limits
- Optimize resource usage
- Profile and optimize code

### Database Scaling

- Read replicas for queries
- Sharding for large datasets
- Connection pooling
- Query optimization

### Cache Scaling

- Redis cluster mode
- Consistent hashing
- Cache warming strategies

## Backup & Recovery

### Database Backup

```bash
# Daily backup
mysqldump -u root -p voucher_db > backup_$(date +%Y%m%d).sql

# Restore
mysql -u root -p voucher_db < backup_20240115.sql
```

### Disaster Recovery

1. Regular backups (daily)
2. Off-site storage (S3)
3. Test recovery procedures monthly
4. Document recovery steps
5. RTO: 1 hour, RPO: 15 minutes

## Troubleshooting

### Common Issues

**Service won't start:**
```bash
# Check logs
kubectl logs -f deployment/voucher-service

# Check events
kubectl describe pod <pod-name>
```

**Database connection issues:**
```bash
# Test connection
mysql -h mysql-service -u voucheruser -p

# Check network policies
kubectl get networkpolicies
```

**High memory usage:**
```bash
# Profile memory
curl http://localhost:8080/debug/pprof/heap > heap.prof
go tool pprof heap.prof
```

## Support

For production issues:
1. Check logs and metrics
2. Review recent deployments
3. Check infrastructure status
4. Contact on-call engineer

