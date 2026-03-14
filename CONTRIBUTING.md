# Contributing to Biblium

Contributions are welcome! Here's how to get started.

## Development setup

```bash
git clone https://github.com/mipsou/mcp-biblium.git
cd mcp-biblium
make build
make test
```

## Workflow

1. Fork the repo and create a feature branch from `dev`
2. Make your changes
3. Run `make test` and `make lint` (requires [golangci-lint](https://golangci-lint.run/))
4. Open a pull request against `dev`

## Code style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Keep changes focused — one PR per feature or fix
- Add tests for new functionality

## Reporting bugs

Open an [issue](https://github.com/mipsou/mcp-biblium/issues) with:

- Steps to reproduce
- Expected vs actual behavior
- OS and Go version

## Contact

- Discord: [mipsou](https://discord.gg/QfXnfHW)
- Twitch: [mipsou](https://www.twitch.tv/mipsou)

## Security

For security vulnerabilities, use [GitHub Security Advisories](https://github.com/mipsou/mcp-biblium/security/advisories/new) instead of public issues. See [SECURITY.md](SECURITY.md).

## License

This project is dual-licensed: [EUPL-1.2](LICENSE) OR [BSD-2-Clause](LICENSE-BSD). By contributing, you agree that your contributions will be available under both licenses.
