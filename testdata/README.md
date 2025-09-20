# Test Data Directory

This directory contains test fixtures and data used by the test suite.

## Organization

- `fixtures/` - Static test data files (JSON, YAML, etc.)
- `mocks/` - Mock data and configurations
- `scenarios/` - Complex test scenarios with multiple files

## Usage

Test files can reference data in this directory using relative paths:

```go
data, err := os.ReadFile("../../testdata/fixtures/sample.json")
```

## Adding Test Data

1. Create appropriately named subdirectories
2. Use descriptive filenames
3. Include comments in test files explaining the test data purpose
4. Keep files small and focused on specific test scenarios