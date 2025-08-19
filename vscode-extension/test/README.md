# M2E VSCode Extension Test Suite

This directory contains the comprehensive test suite for the M2E VSCode extension.

## Test Structure

```
test/
├── README.md                    # This file
├── runTest.ts                   # Main test runner
├── .mocharc.json               # Mocha configuration
├── suite/
│   └── index.ts                # Test suite entry point
├── unit/                       # Unit tests
│   ├── utils.test.ts           # Utility function tests
│   ├── textProcessing.test.ts  # Text processing tests
│   └── client.test.ts          # API client tests
├── integration/                # Integration tests
│   ├── extension.test.ts       # Extension integration tests
│   └── diagnostics.test.ts     # Diagnostics provider tests
├── helpers/                    # Test helpers and utilities
│   ├── mockServer.ts           # Mock M2E server
│   └── testUtils.ts            # Test utility functions
└── fixtures/                   # Test data and fixtures
    ├── testTexts.ts            # Sample text data
    ├── mockResponses.ts        # Mock API responses
    └── sampleFiles.ts          # Sample file contents
```

## Running Tests

### All Tests
```bash
npm test
```

### Unit Tests Only
```bash
npm run test:unit
```

### Integration Tests Only
```bash
npm run test:integration
```

### Watch Mode
```bash
npm run watch:test
```

### From VSCode
- Open Command Palette (`Cmd+Shift+P` / `Ctrl+Shift+P`)
- Run "Tasks: Run Task"
- Select test task to run

## Test Categories

### Unit Tests

Unit tests focus on individual functions and components in isolation:

#### `utils.test.ts`
Tests for core utility functions:
- `isTextFile()` - File type detection
- `getFileType()` - Language detection from file extensions
- `formatBytes()` - File size formatting
- `debounce()` - Function debouncing
- `isLargeSelection()` - Large text detection

#### `textProcessing.test.ts`
Tests for text processing utilities:
- `extractWordAtPosition()` - Word extraction from text position
- `isAmericanSpelling()` - American spelling detection
- `shouldExcludeFromDiagnostics()` - File exclusion logic

#### `client.test.ts`
Tests for API client functionality:
- `getFileTypeFromDocument()` - Document type detection
- Server communication methods
- Error handling scenarios
- Response validation

### Integration Tests

Integration tests verify component interactions and full workflows:

#### `extension.test.ts`
Tests for extension-wide functionality:
- Extension activation and command registration
- Convert Selection command
- Convert File command
- Convert Comments Only command
- Server management
- Error handling
- Configuration changes

#### `diagnostics.test.ts`
Tests for diagnostic provider functionality:
- American spelling detection in various file types
- Diagnostic severity levels
- Real-time updates
- Quick Fix code actions
- Performance with large files
- Configuration integration

## Test Helpers

### MockM2EServer
A complete mock implementation of the M2E server for testing:

```typescript
const mockServer = new MockM2EServer(18182);
await mockServer.start();

// Set custom responses
mockServer.setResponse('/api/v1/convert', customResponse);

await mockServer.stop();
```

Features:
- HTTP server implementation
- Configurable responses
- Request/response logging
- Error simulation

### Test Utilities
Helper functions for VSCode extension testing:

```typescript
// Create test documents
const doc = await createTestDocument('test content', 'javascript');

// Setup test workspace
await setupTestWorkspace({ enableDiagnostics: true });

// Wait for diagnostics
const diagnostics = await waitForDiagnostics(document, 2);

// Execute commands
await executeCommand('m2e.convertSelection');
```

## Test Fixtures

### Test Texts (`testTexts.ts`)
Predefined text samples for various scenarios:
- Simple American/British spelling pairs
- Code samples with comments
- Large text samples
- Edge cases (empty, Unicode, etc.)

### Mock Responses (`mockResponses.ts`)
Predefined API responses:
- Successful conversions
- Error responses
- Response generation utilities

### Sample Files (`sampleFiles.ts`)
Complete file contents for different programming languages:
- JavaScript/TypeScript
- Python
- Go
- Markdown
- Configuration files

## Writing New Tests

### Unit Test Pattern
```typescript
import * as assert from 'assert';
import { functionToTest } from '../../src/utils';

suite('Feature Test Suite', () => {
    test('should behave correctly', () => {
        const result = functionToTest('input');
        assert.strictEqual(result, 'expected');
    });
});
```

### Integration Test Pattern
```typescript
import * as vscode from 'vscode';
import { createTestDocument, executeCommand } from '../helpers/testUtils';

suite('Integration Test Suite', () => {
    test('should integrate correctly', async () => {
        const document = await createTestDocument('test content');
        await executeCommand('m2e.convertSelection');
        
        const updatedText = document.getText();
        assert.ok(updatedText.includes('expected'));
    });
});
```

### Test Best Practices

1. **Isolation**: Each test should be independent
2. **Cleanup**: Clean up resources after tests
3. **Assertions**: Use descriptive assertion messages
4. **Async**: Properly handle async operations
5. **Mocking**: Use mocks for external dependencies

## Debugging Tests

### VSCode Debugging
1. Set breakpoints in test files
2. Use "Extension Tests" launch configuration
3. Step through code in Debug Console

### Console Logging
```typescript
import { MockConsole } from '../helpers/testUtils';

const mockConsole = new MockConsole();
mockConsole.start();

// Run test code

assert.ok(mockConsole.hasLog('expected message'));
mockConsole.stop();
```

### Test Output
Check the Test Results panel in VSCode for detailed test output and failure messages.

## Performance Testing

### Large File Tests
Tests verify performance with large text files:
- 100KB+ text processing
- Multiple document handling
- Memory usage monitoring

### Diagnostic Performance
Tests for diagnostic provider efficiency:
- Real-time updates
- Debounced processing
- Large workspace handling

## Continuous Integration

Tests run automatically in CI/CD:
- On pull requests
- Before merging to main
- During release process

### CI Configuration
See `.github/workflows/test.yml` for CI setup.

## Coverage Reporting

Generate test coverage reports:
```bash
npm run test:coverage
```

Coverage reports show:
- Line coverage
- Function coverage
- Branch coverage
- Uncovered code paths

## Troubleshooting

### Common Issues

#### Tests Timing Out
- Increase timeout in `.mocharc.json`
- Check for unresolved promises
- Verify mock server is running

#### Extension Not Activating
- Check activation events
- Verify extension is built (`npm run compile`)
- Check for registration errors

#### Mock Server Issues
- Ensure unique ports for parallel tests
- Check server startup/shutdown logic
- Verify response configuration

#### VSCode API Issues
- Check VSCode API version compatibility
- Verify test environment setup
- Use proper async/await patterns

### Getting Help

1. Check test output for specific errors
2. Review test documentation
3. Check existing similar tests
4. Run tests in isolation to identify issues

---

For more information about the M2E extension, see the [main README](../README.md).