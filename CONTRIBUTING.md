# Contributing to OwlRelay

Thank you for your interest in contributing! 🦉

## Getting Started

1. Fork the repository
2. Clone your fork
3. Create a feature branch: `git checkout -b feature/my-feature`
4. Make your changes
5. Run tests: `make test`
6. Commit: `git commit -m "feat: add my feature"`
7. Push: `git push origin feature/my-feature`
8. Open a Pull Request

## Development Setup

### Relay Server (Go)

```bash
cd relay
go mod download
go run ./cmd/relay serve
```

### Extension (TypeScript)

```bash
cd extension
npm install
npm run dev
```

## Code Style

- **Go**: Follow standard Go conventions, use `gofmt`
- **TypeScript**: Follow ESLint config
- **Commits**: Use [Conventional Commits](https://conventionalcommits.org)
  - `feat:` new feature
  - `fix:` bug fix
  - `docs:` documentation
  - `refactor:` code refactoring
  - `test:` adding tests

## Pull Request Guidelines

- Keep PRs focused on a single change
- Include tests for new features
- Update documentation if needed
- Ensure CI passes

## Reporting Issues

- Search existing issues first
- Include reproduction steps
- Include browser/OS version
- Include relay server logs if relevant

## Security

If you discover a security vulnerability, please email emre@emreyilmaz.io instead of opening a public issue.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
