# Contributing to Pulumi Kind Cluster

Thank you for your interest in contributing! This document provides guidelines and instructions for contributing to this project.

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment for everyone.

## How to Contribute

### Reporting Bugs

If you find a bug, please create an issue with:

1. **Clear title** - Summarize the issue
2. **Description** - Detailed explanation of the bug
3. **Steps to reproduce** - How to recreate the issue
4. **Expected behavior** - What should happen
5. **Actual behavior** - What actually happens
6. **Environment details** - OS version, Pulumi version, Go version, etc.
7. **Logs** - Relevant error messages or stack traces

### Suggesting Enhancements

For feature requests or enhancements:

1. **Check existing issues** - Avoid duplicates
2. **Provide use case** - Explain why this would be useful
3. **Describe the solution** - How you envision it working
4. **Consider alternatives** - Other approaches you've thought about

### Pull Requests

1. **Fork the repository**
   ```bash
   gh repo fork tstark7952/pulumi-kind-cluster --clone
   cd pulumi-kind-cluster
   ```

2. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Make your changes**
   - Write clean, readable code
   - Follow Go best practices
   - Add comments for complex logic
   - Update documentation as needed

4. **Test your changes**
   ```bash
   # Run Pulumi preview
   pulumi preview

   # If safe, test deployment
   pulumi up

   # Verify functionality
   kubectl get nodes

   # Clean up
   pulumi destroy
   ```

5. **Commit your changes**
   ```bash
   git add .
   git commit -m "feat: add your feature description"
   ```

   Use conventional commit messages:
   - `feat:` - New feature
   - `fix:` - Bug fix
   - `docs:` - Documentation changes
   - `refactor:` - Code refactoring
   - `test:` - Adding tests
   - `chore:` - Maintenance tasks

6. **Push to your fork**
   ```bash
   git push origin feature/your-feature-name
   ```

7. **Create a Pull Request**
   - Use a clear title and description
   - Reference any related issues
   - Explain what changed and why
   - Include screenshots if applicable

## Development Guidelines

### Code Style

- Follow standard Go formatting (`gofmt`)
- Use meaningful variable and function names
- Keep functions focused and concise
- Add comments for exported functions

### Testing

Before submitting a PR:

1. **Test locally** - Deploy and verify the cluster works
2. **Test cleanup** - Ensure `pulumi destroy` removes everything
3. **Test different configurations** - Try various CPU/memory settings
4. **Check error handling** - Verify proper error messages

### Documentation

Update documentation when you:

- Add new configuration options
- Change existing behavior
- Add new features
- Fix bugs that affected documented behavior

## Project Structure

```
.
├── main.go              # Main Pulumi program
├── go.mod               # Go dependencies
├── Pulumi.yaml          # Pulumi project config
├── README.md            # Main documentation
├── CONTRIBUTING.md      # This file
├── LICENSE              # MIT license
└── .gitignore           # Git ignore rules
```

## Development Setup

1. **Install dependencies**
   ```bash
   brew install pulumi go lima kind kubectl
   ```

2. **Clone the repository**
   ```bash
   git clone https://github.com/tstark7952/pulumi-kind-cluster.git
   cd pulumi-kind-cluster
   ```

3. **Initialize Pulumi stack**
   ```bash
   pulumi stack init dev
   ```

4. **Make changes and test**
   ```bash
   pulumi preview
   pulumi up
   ```

## Review Process

1. Maintainers will review your PR
2. Address any requested changes
3. Once approved, your PR will be merged
4. Your contribution will be credited

## Questions?

If you have questions:

- Open an issue for discussion
- Check existing issues and PRs
- Review the README for documentation

## Recognition

Contributors will be recognized in the project. Thank you for helping improve this project!

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
