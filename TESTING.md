# Test Suite for Xplorer

## Overview
Comprehensive test suite covering core functionality of the Xplorer file explorer project (command: `xp`).

## Test Coverage

### ✅ Feature Tests

#### Path Toggle Feature (New)
- **TestPathToggle**: Validates the new `showRawPath` toggle functionality
  - Tests initial state (false)
  - Tests toggle on (true)
  - Tests toggle off (false)
  - Verifies the `p` key binding works correctly

#### File System Operations
- **TestFileIcon**: Validates file type icons for different extensions
  - Directories, Go, Python, JS, TS, JSON, HTML, etc.
  - Ensures icons are returned for all file types

- **TestFileColor**: Tests color mapping for files
  - Directory colors
  - Extension-based colors (.go, .py, .js)
  - Default color for unknown types

#### Utility Functions
- **TestFormatSize**: File size formatting
  - Bytes (512 B)
  - Kilobytes (1.0 KB)
  - Megabytes (1.0 MB)
  - Gigabytes (1.0 GB)

- **TestBoolStr**: Boolean to string conversion
  - true → "ON"
  - false → "OFF"

- **TestMinMax**: Min/max utility functions
  - Positive numbers
  - Negative numbers
  - Equal values

#### Bookmark Features
- **TestIsBookmarked**: Bookmark detection
  - Exact path matches
  - Non-bookmarked paths
  - Paths with trailing slashes

#### Language Detection
- **TestDetectLanguage**: Syntax highlighting language detection
  - 14 different languages tested
  - Go, Python, JS, TS, C/C++, Java, Ruby, Rust, PHP, HTML, CSS, JSON, Shell
  - Unknown file types return empty string

#### File Type Description
- **TestDescribeFileByExt**: Human-readable file descriptions
  - Executables (EXE, DLL)
  - Images (PNG, JPG)
  - Archives (ZIP)
  - Documents (PDF)
  - Media (MP4, MP3)
  - Binary files

#### Path Operations
- **TestPathCleaning**: Path normalization
  - Clean paths
  - Trailing slashes
  - Dot notation (./)
  - Parent directory (..)

#### Configuration
- **TestConfigDefaults**: Validates default keybindings
  - Filter key: '/'
  - Toggle path key: 'p'
  - Quit key, help key, etc.

### 🚀 Benchmark Tests

- **BenchmarkFileIcon**: Icon lookup performance
- **BenchmarkFileColor**: Color lookup performance
- **BenchmarkFormatSize**: Size formatting performance
- **BenchmarkDetectLanguage**: Language detection performance

## Running Tests

### Run all tests
```bash
go test -v
```

### Run specific test
```bash
go test -v -run TestPathToggle
```

### Run benchmarks
```bash
go test -bench=.
```

### Run with coverage
```bash
go test -cover
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Test Results

All tests passing ✅:
- 12 test functions
- 70+ individual test cases
- 4 benchmark functions

```
PASS
ok      github.com/alexcostache/Xplorer/tests      0.745s
```

## New Feature Validation

The path toggle feature (`p` key) has been successfully tested:
1. ✅ Variable `showRawPath` toggles correctly
2. ✅ Key binding `togglePathKey = 'p'` is configured
3. ✅ Integration with drawAddressBar function
4. ✅ Display modes work:
   - Raw path: `Path: /Users/username/projects/xp`
   - Breadcrumb: `~ › projects › xp`

## CI/CD Integration

Add to your workflow:
```yaml
- name: Run tests
  run: go test -v ./...

- name: Run benchmarks
  run: go test -bench=. -benchmem
```

---

*Last updated: December 21, 2025*
