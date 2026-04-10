# Contributing to UPSWake

Firstly, thank you for considering contributing to UPSWake! We welcome contributions from the community to help improve
and enhance the project.

## Security Reports

If you discover a security vulnerability in UPSWake, please do not open a pull request or issue. Instead, please follow
the instructions in our [Security Policy](https://github.com/TheDarthMole/upswake/security/policy) to report the
vulnerability privately.

## Raise an Issue

Before starting work on a new feature or bug fix, please check the issue tracker to see if someone else has already
reported the same issue. If not, please open a new issue to discuss your proposed changes with any relevant information
required for context (including configuration details, logs, and steps to reproduce the issue). Please redact any
sensitive information such as passwords or public IP addresses.

## Make a Pull Request

If you have a fix or feature ready, please fork the repository and create a pull request.
Once you have made a pull request, a maintainer will review your changes and decide what to do next.

Please ensure that you follow these guidelines:

### Check the license

By contributing, you agree that your contributions will be licensed under the same license as the
project, and you assert that you have full power to license your contribution under the [GPL-3.0 License](../LICENSE).
Please refer to the [LICENSE](../LICENSE) file for more details.

### Write clear commit messages

Please write clear and concise commit messages that describe the changes you have made. Generally following the
[Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/#summary) specification is a good idea.

### Setting up your development environment

This project uses [Golang](https://go.dev/), [Just](https://github.com/casey/just), [mise](https://mise.jdx.dev/) and your choice of [Docker](https://www.docker.com/) or [Podman](https://podman.io/) for development.
To install the dependencies, please run the following commands:

```bash
# Install 'mise' from https://mise.jdx.dev/
# Also install podman or docker from their respective websites
mise install # Install the rest of the dependencies
hk install # Install hk pre-commit checker
```

Just is used to build or run the application along with other useful development commands, you can use the `just` tool
to run the commands defined in the [justfile](../justfile).

```text
just -l
Available recipes:
    build            # Build upswake
    build-container  # Build the thedarthmole/upswake:local container
    clean            # Clean up generated files and test cache
    fmt              # Runs all formatters
    generate-certs   # Generate self-signed certificates for testing
    help             # Display this help message
    install-deps     # Install development dependencies
    lint             # Runs all linters
    run *args        # Run upswake with arguments
    run-container    # Builds and runs the upswake container
    start-nut-server # Runs a NUT server in a container for testing
    stop-container   # Stops the upswake container
    stop-nut-server  # Stops the NUT server container
    test             # Run all Go tests, assuming the NUT server is already running and certs are generated
    test-local       # Run all Go tests locally
```

### Follow the coding style

Please follow the existing coding style and conventions used in the project. [Hk](https://hk.jdx.dev/) is used as a pre-commit checker to
ensure standards for the project are maintained. Ensure that all checks pass before submitting your pull request.
You can run the linters using:

```bash
just lint
just fmt # to automatically fix any issues that can be fixed automatically
```

### Write tests

Please write tests for any new features or bug fixes you have implemented. Ensure that all tests pass before submitting
your pull request. You can run the tests using:

```bash
just test
```

### Update documentation

Please update the documentation to reflect any changes you have made. This includes updating the README.md file as well
as the swagger documentation for any API changes.

## Maintainers

[TheDarthMole](mailto:upswake@darthmole.dev) is currently the only maintainer of UPSWake and makes all final decisions.
