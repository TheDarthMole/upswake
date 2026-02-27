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

### Follow the coding style

Please follow the existing coding style and conventions used in the project. Golang-ci lint 
is used for linting, ensure that all checks pass before submitting your pull request. You can run the linters using:

```bash
just install-deps # Install golangci-lint and swag if you haven't already
just lint
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
