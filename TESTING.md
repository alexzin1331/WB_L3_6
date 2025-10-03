# Testing Guide

This project includes comprehensive tests for the storage layer using testcontainers for PostgreSQL integration testing.

## Prerequisites

- Go 1.21+
- Docker (required for testcontainers)
- Docker Compose (optional, for manual testing)

## Test Setup

### Automatic Setup

Run the Make targets to automatically download dependencies and run tests:

```bash
# Install all dependencies
make deps

# Install test-specific dependencies
make test-deps

# Run all tests
make test

# Run tests with coverage
make test-coverage
```

### Manual Setup

1. **Install dependencies:**
   ```bash
   go mod download
   go mod tidy
   ```

2. **Run tests:**
   ```bash
   # Run all tests
   go test ./internal/storage/

   # Run tests with verbose output
   go test -v ./internal/storage/

   # Run tests with coverage
   go test -coverprofile=coverage.out ./internal/storage/
   ```

## Test Structure

### Test Files

- `internal/storage/storage_test.go` - Comprehensive test suite for storage operations

### Test Categories

1. **CRUD Operations Tests**
   - `TestStorage_CreateSale` - Tests sale creation with validation
   - `TestStorage_GetSales` - Tests retrieving sales with ordering
   - `TestStorage_UpdateSale` - Tests updating existing sales
   - `TestStorage_DeleteSale` - Tests deleting sales

2. **Analytics Tests**
   - `TestStorage_GetAnalytics` - Tests statistical calculations (sum, average, median, percentiles)

3. **Error Handling Tests**
   - `TestStorage_ErrorHandling` - Tests database constraints and error scenarios

4. **Performance Tests**
   - `BenchmarkStorage_CreateSale` - Benchmarks sale creation
   - `BenchmarkStorage_GetSales` - Benchmarks sale retrieval

### Test Features

- **Testcontainers Integration**: Uses PostgreSQL containers for real database testing
- **Automatic Migration**: Schema is set up automatically for each test
- **Data Isolation**: Each test runs in a clean database state
- **Realistic Test Data**: Uses diverse sample data covering edge cases
- **Constraint Validation**: Tests database-level constraints
- **Performance Benchmarking**: Includes benchmarks for critical operations

## Test Data

The tests use predefined sample data covering:

- Income transactions (Salary, Freelance)
- Expense transactions (Food, Rent)
- Various amounts and dates
- Different categories
- Edge cases for testing constraints

## Database Schema

The tests automatically create the following schema:

```sql
CREATE TABLE sales (
    id SERIAL PRIMARY KEY,
    type VARCHAR(10) NOT NULL CHECK (type IN ('income', 'expense')),
    amount DECIMAL(10,2) NOT NULL CHECK (amount > 0),
    date TIMESTAMPTZ NOT NULL,
    category VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sales_date ON sales(date);
CREATE INDEX idx_sales_category ON sales(category);
```

## Running Specific Tests

```bash
# Run only CRUD tests
go test -run TestStorage_CreateSale ./internal/storage/

# Run only analytics tests
go test -run TestStorage_GetAnalytics ./internal/storage/

# Run only error handling tests
go test -run TestStorage_ErrorHandling ./internal/storage/

# Run benchmarks
go test -bench=. ./internal/storage/
```

## Coverage Reports

Generate and view coverage reports:

```bash
# Generate coverage report
make test-coverage

# Open coverage report in browser
open coverage.html
```

## Troubleshooting

### Docker Issues

If you encounter Docker-related issues:

1. Ensure Docker is running
2. Check Docker Desktop resources if on macOS/Windows
3. Verify Docker Compose compatibility

### Test Timeouts

If tests timeout:

1. Check Docker container resources
2. Ensure sufficient disk space
3. Verify network connectivity

### Database Connection Issues

If PostgreSQL connection fails:

1. Verify Docker container is running
2. Check container logs
3. Ensure port conflicts don't exist

## Continuous Integration

These tests are designed to work in CI environments:

- No persistent state requirements
- Automatic cleanup
- Docker container reuse
- Parallel test execution

Example CI configuration:

```yaml
test:
  script:
    - make deps
    - make test-coverage
  services:
  - docker:dind
```

## Performance Considerations

- Tests use lightweight PostgreSQL containers
- Containers are reused across test functions
- Parallel benchmark execution
- Memory-efficient test data cleanup

## Best Practices Demonstrated

1. **Isolation**: Each test is independent
2. **Realism**: Tests use actual Docker containers
3. **Coverage**: Comprehensive test scenarios
4. **Performance**: Includes benchmarks
5. **Maintainability**: Clear test structure and naming
6. **Reliability**: Proper error handling and cleanup
