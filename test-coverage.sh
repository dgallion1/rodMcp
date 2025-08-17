#!/bin/bash

# This script generates a test coverage report for the Go project.

set -e

# List of packages to include in test coverage
PACKAGES=$(go list ./... | grep -v /examples | grep -v /old_tests)

# Generate a coverage report for all packages.
go test -coverprofile=coverage.out $PACKAGES

# Display the coverage percentage.
go tool cover -func=coverage.out

# Generate an HTML report for detailed analysis.
go tool cover -html=coverage.out -o coverage.html

echo "Coverage report generated: coverage.html"
