# Contributing to RiskEngine

Thank you for your interest in contributing! This document covers how to get started, coding conventions, and the PR process.

## Getting Started

1. **Fork** the repository and clone your fork locally.
2. Install prerequisites: Go 1.22+, Docker, `golangci-lint`, `mockery`.
3. Run `make setup` to install all tooling.
4. Create a feature branch: `git checkout -b feat/my-feature`.

## Development Workflow

```bash
make setup    # install tools
make test     # unit tests (requires Redis; use docker compose dev)
make lint     # golangci-lint
make bench    # benchmarks
make proto    # regenerate protobuf (requires protoc)
```

Start local dependencies:

```bash
docker compose -f deployments/docker/compose.dev.yaml up -d
```

## Code Conventions

- Follow the rules in `.cursor/rules/` (go-standards, riskengine-domain, performance).
- All exported types and functions must have godoc comments.
- New packages need a `doc.go` with a package-level comment.
- Error strings: lowercase, no period at the end.
- No `panic` in library code; return errors.

## Adding a Rule / Feature / Model

See the step-by-step guides in `docs/`:
- [docs/adding-rules.md](docs/adding-rules.md)
- [docs/adding-features.md](docs/adding-features.md)
- [docs/adding-models.md](docs/adding-models.md)

## Testing Requirements

- Unit tests required for all new `internal/` code.
- Benchmark (`BenchmarkXxx`) required for hot-path code (rule evaluators, feature fetchers).
- Integration tests go in `internal/<pkg>/testdata/` and use `testcontainers-go` for Redis/Kafka.
- Aim for > 80% coverage on new code (CI enforces this).

## Pull Request Checklist

- [ ] `make test` passes locally
- [ ] `make lint` passes with zero warnings
- [ ] Benchmarks added for hot-path changes
- [ ] `CHANGELOG.md` updated (use [Keep a Changelog](https://keepachangelog.com) format)
- [ ] PR description explains **why** (not just what)
- [ ] Breaking changes noted with `BREAKING:` prefix in commit message

## Commit Message Format

We follow [Conventional Commits](https://www.conventionalcommits.org):

```
<type>(<scope>): <short description>

[optional body]

[optional footer]
```

Types: `feat`, `fix`, `perf`, `refactor`, `test`, `docs`, `chore`, `ci`

Examples:
```
feat(rule): add velocity-based device multi-account rule
perf(feature): parallelize Redis and HBase fetchers
fix(list): fix Bloom filter false-positive on rehash
```

## Security Issues

Do **not** open public issues for security vulnerabilities. Email `security@yourorg.com` instead. We follow responsible disclosure and will respond within 48 hours.

## License

By contributing, you agree that your contributions will be licensed under the Apache 2.0 License.
