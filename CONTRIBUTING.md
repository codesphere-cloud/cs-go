# Contributing to Codesphere Go SDK & CLI

We welcome contributions of all kinds! By participating in this project, you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md).

## How to Report Issues

If you encounter a bug or have a feature request, please [open a new issue](https://github.com/codesphere-cloud/cs-go/issues/new) on GitHub. Please include the following information:

* **Operating System and Version:**
* **Codesphere Go SDK & CLI Version (if applicable):**
* **Steps to Reproduce the Bug:**
* **Expected Behavior:**
* **Actual Behavior:**
* **Any relevant logs or error messages:**

## How to Suggest Features or Improvements

We'd love to hear your ideas! Please [open a new issue](https://github.com/codesphere-cloud/cs-go/issues/new) to discuss your proposed feature or improvement before submitting code. This allows us to align on the design and approach.

## Contributing Code

If you'd like to contribute code, please follow these steps:

1.  **Fork the Repository:** Fork this repository to your GitHub account.
2.  **Create a Branch:** Create a new branch for your changes: `git checkout -b feature/your-feature-name`
3.  **Set Up Development Environment:**

    * Ensure you have Go installed.  The minimum required Go version is specified in the `go.mod` file.
    * Clone your forked repository: `git clone git@github.com:your-username/cs-go.git`
    * Navigate to the project directory: `cd cs-go`
    * Run `make`: This command should download necessary dependencies and build the CLI.

4.  **Follow Coding Standards:**

    * Please ensure your code is properly formatted using `go fmt`.
    * We use [golangci-lint](https://golangci-lint.run/) for static code analysis. Please run it locally before submitting a pull request: `make lint`.

5.  **Write Tests:**

    * We use [Ginkgo](https://github.com/onsi/ginkgo) and [Gomega](https://github.com/onsi/gomega) for testing.
    * Please write tests for your code using Ginkgo and Gomega and add them to the `_test.go` files.
    * Aim for good test coverage.

6.  **Build and Test:**

    * Ensure everything is working correctly by running the appropriate `make` targets (e.g., `make build`, `make test`). The `make test` target should run the Ginkgo tests.

7.  **Commit Your Changes:**

    * We use [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) for our commit messages. Please format your commit messages according to the Conventional Commits specification. Examples:
        * `fix(api): Handle edge case in API client`
        * `feat(cli): Add new command for listing resources`
        * `docs: Update contributing guide with commit message conventions`

8.  **Submit a Pull Request:** [Open a new pull request](https://github.com/codesphere-cloud/cs-go/compare) to the `main` branch of this repository. Please include a clear description of your changes and reference any related issues.

## Code Review Process

All contributions will be reviewed by project maintainers. Please be patient during the review process and be prepared to make revisions based on feedback. We aim for thorough but timely reviews.

## License

By contributing to Codesphere Go SDK & CLI, you agree that your contributions will be licensed under the [Apache License 2.0](LICENSE).

## Community

Connect with the community and ask questions by joining our mailing list: [cs-go@codesphere.com](mailto:cs-go@codesphere.com).

Thank you for your interest in contributing to Codesphere Go SDK & CLI!
