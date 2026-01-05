# Documentation Index

Welcome to the Voucher & Payment Service documentation! This index helps you navigate to the right documentation for your needs.

## 🚀 Getting Started

**New to the project?** Start here:

1. **[QUICKSTART.md](QUICKSTART.md)** - Get running in 5 minutes
   - Prerequisites
   - Automated setup
   - First API call
   - Troubleshooting

2. **[README.md](README.md)** - Complete project documentation
   - Overview & features
   - Architecture diagram
   - Installation guide
   - API usage examples
   - Development workflow

## 📚 Core Documentation

### For Developers

**[ARCHITECTURE.md](ARCHITECTURE.md)** - System design deep-dive
- Clean architecture explanation
- Component design
- Data flow diagrams
- Design decisions
- Scalability strategies

**[CONTRIBUTING.md](CONTRIBUTING.md)** - Development guidelines
- Code style guide
- Testing standards
- Pull request process
- Commit conventions
- Security guidelines

### For Users

**[API_GUIDE.md](API_GUIDE.md)** - Complete API reference
- All endpoints documented
- Request/response examples
- Error handling
- Authentication
- Best practices

**[DEPLOYMENT.md](DEPLOYMENT.md)** - Production deployment
- Docker deployment
- Kubernetes deployment
- Production checklist
- Monitoring setup
- Troubleshooting

### Project Overview

**[PROJECT_SUMMARY.md](PROJECT_SUMMARY.md)** - High-level overview
- Project statistics
- Technology stack
- Key features
- Learning objectives
- Production readiness

## 📖 Documentation by Role

### I want to...

#### ...understand what this project is
→ Start with [README.md](README.md) Overview section
→ Then read [PROJECT_SUMMARY.md](PROJECT_SUMMARY.md)

#### ...get it running quickly
→ Follow [QUICKSTART.md](QUICKSTART.md)
→ Run the demo: `./scripts/demo.sh`

#### ...understand the architecture
→ Read [ARCHITECTURE.md](ARCHITECTURE.md)
→ Review code structure in [README.md](README.md)

#### ...use the API
→ Check [API_GUIDE.md](API_GUIDE.md)
→ Try examples in [QUICKSTART.md](QUICKSTART.md)

#### ...contribute code
→ Read [CONTRIBUTING.md](CONTRIBUTING.md)
→ Review coding standards
→ Check project structure

#### ...deploy to production
→ Follow [DEPLOYMENT.md](DEPLOYMENT.md)
→ Review production checklist
→ Setup monitoring

#### ...debug an issue
→ Check [DEPLOYMENT.md](DEPLOYMENT.md) Troubleshooting
→ Review logs in Jaeger: http://localhost:16686
→ Check [ARCHITECTURE.md](ARCHITECTURE.md) for system design

## 🎯 Documentation by Topic

### Setup & Installation
- [QUICKSTART.md](QUICKSTART.md) - Quick setup guide
- [README.md](README.md) - Detailed installation
- `scripts/setup.sh` - Automated setup (Linux/macOS)
- `scripts/setup.bat` - Automated setup (Windows)

### Architecture & Design
- [ARCHITECTURE.md](ARCHITECTURE.md) - System architecture
- [README.md](README.md) - Architecture overview
- [PROJECT_SUMMARY.md](PROJECT_SUMMARY.md) - Technology stack

### API Documentation
- [API_GUIDE.md](API_GUIDE.md) - Complete API reference
- [README.md](README.md) - API overview
- `api/proto/voucher.proto` - Protocol definitions

### Development
- [CONTRIBUTING.md](CONTRIBUTING.md) - Development guide
- [README.md](README.md) - Development section
- `Makefile` - Build commands

### Testing
- [CONTRIBUTING.md](CONTRIBUTING.md) - Testing guidelines
- [README.md](README.md) - Testing section
- `scripts/test.sh` - Test runner

### Deployment
- [DEPLOYMENT.md](DEPLOYMENT.md) - Full deployment guide
- [README.md](README.md) - Deployment section
- `docker-compose.yml` - Infrastructure setup
- `Dockerfile` - Container image

### Operations
- [DEPLOYMENT.md](DEPLOYMENT.md) - Monitoring & troubleshooting
- [README.md](README.md) - Observability section

## 📂 Code Documentation

### Entry Points
- `cmd/server/main.go` - Application entry point
- `cmd/migrate/main.go` - Database migrations

### Core Components
- `internal/domain/` - Business entities and rules
- `internal/service/` - Business logic
- `internal/repository/` - Data access layer
- `internal/grpc/` - API handlers

### Infrastructure
- `internal/cache/` - Redis caching
- `internal/messaging/` - RabbitMQ integration
- `internal/observability/` - Tracing & logging
- `internal/worker/` - Worker pools

## 🔧 Configuration Files

- `go.mod` - Go dependencies
- `docker-compose.yml` - Infrastructure services
- `config/config.yaml` - Application configuration
- `config/prometheus.yml` - Metrics configuration
- `Makefile` - Build automation
- `.gitignore` - Git exclusions
- `.dockerignore` - Docker exclusions

## 📜 SQL & Migrations

- `migrations/init.sql` - Database schema
  - All table definitions
  - Indexes and constraints
  - Sample data

## 🛠️ Helper Scripts

- `scripts/setup.sh` - Setup for Linux/macOS
- `scripts/setup.bat` - Setup for Windows
- `scripts/test.sh` - Run all tests
- `scripts/demo.sh` - Interactive demo

## 📊 Generated Files

- `api/generated/` - Generated gRPC code
  - `*.pb.go` - Protocol Buffer code
  - `*.pb.gw.go` - Gateway code

## 🎓 Learning Path

Recommended reading order for learning:

1. **Day 1**: [QUICKSTART.md](QUICKSTART.md) → Get it running
2. **Day 2**: [README.md](README.md) → Understand the project
3. **Day 3**: [ARCHITECTURE.md](ARCHITECTURE.md) → Learn the design
4. **Day 4**: [API_GUIDE.md](API_GUIDE.md) → Use the API
5. **Day 5**: Code exploration → Read the implementation
6. **Day 6**: [CONTRIBUTING.md](CONTRIBUTING.md) → Development practices
7. **Day 7**: [DEPLOYMENT.md](DEPLOYMENT.md) → Production deployment

## 🆘 Getting Help

### Troubleshooting
1. Check [QUICKSTART.md](QUICKSTART.md) Troubleshooting section
2. Review [DEPLOYMENT.md](DEPLOYMENT.md) Troubleshooting section
3. Check logs: `docker-compose logs`
4. Search closed issues on GitHub

### Common Questions

**Q: How do I start the service?**
A: See [QUICKSTART.md](QUICKSTART.md) Step 3

**Q: What are the API endpoints?**
A: Check [API_GUIDE.md](API_GUIDE.md)

**Q: How do I contribute?**
A: Read [CONTRIBUTING.md](CONTRIBUTING.md)

**Q: How do I deploy to production?**
A: Follow [DEPLOYMENT.md](DEPLOYMENT.md)

**Q: Where is the architecture explained?**
A: See [ARCHITECTURE.md](ARCHITECTURE.md)

## 📞 Support

- **Documentation Issues**: Check this index
- **Code Issues**: See [CONTRIBUTING.md](CONTRIBUTING.md)
- **Deployment Issues**: See [DEPLOYMENT.md](DEPLOYMENT.md)
- **API Questions**: See [API_GUIDE.md](API_GUIDE.md)

## 📝 Document Versions

All documentation is version controlled with the code.
- Latest version: In repository
- Specific version: Check git tags

## 🔄 Keeping Documentation Updated

When making changes:
1. Update relevant documentation files
2. Update this index if adding new docs
3. Mention doc changes in PR
4. Keep code and docs in sync

## ✨ Quick Reference Card

```
Start Service:     make run
Stop Service:      Ctrl+C
Run Tests:         make test
View API Docs:     API_GUIDE.md
View Traces:       http://localhost:16686
View Queue:        http://localhost:15672
gRPC Port:         50051
HTTP Port:         8080
```

---

**Need something not listed here?**
- Check [README.md](README.md) for overview
- Search repository files
- Open an issue on GitHub

**Happy Coding! 🚀**

