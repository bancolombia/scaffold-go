# Contributing

Thanks for your interest in contributing to `scaffold-go`! Bug reports, documentation improvements, and code contributions are welcome.

## Before you start

- For bugs, search existing issues first. Include the command you ran, what you expected, what happened, and the relevant error output.
- For larger changes or new features, open an issue to discuss the approach before investing in an implementation.
- Keep changes focused. When changing generated code, include or update tests that verify the generated output.

## Development setup

The project requires the Go version declared in [`go.mod`](go.mod). Fork the repository, then clone your fork and create a branch:

```sh
git clone https://github.com/<your-account>/scaffold-go.git
cd scaffold-go
git switch -c your-change
```

Run the checks relevant to your change before opening a pull request:

```sh
make test-fast        # Fast tests for development
make test-integration # Integration tests
make test             # Complete test suite, including property tests
make build            # Build the CLI
make fmt              # Format Go code
```

The full test suite can take longer because it includes property-based tests. Please run `make test` when practical; if you cannot, mention which checks you did run in your pull request.

## Pull requests

Open a pull request against the default branch and include:

- A clear summary of the problem and the change.
- The related issue, if there is one.
- The tests and other checks you ran, plus any known limitations.
- Documentation updates when behavior or commands change.

Please keep each pull request focused and respond to review feedback so maintainers can understand and evaluate the change.

## License

By submitting a contribution, you agree that it may be distributed under the project's [Apache License 2.0](LICENCE).
