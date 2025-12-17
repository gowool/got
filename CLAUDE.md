# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**got** is a Go template theme management library that provides a flexible system for organizing and rendering HTML templates with support for template inheritance, themes, and parent-child relationships.

## Core Architecture

### Main Components

1. **Theme (`theme.go`)**: Central theme manager that handles template rendering, caching, and inheritance
   - Supports parent-child theme relationships for template fallback
   - Template caching with debug mode toggles
   - Custom function maps for template execution
   - Automatic template dependency resolution

2. **Template (`template.go`)**: Template interface and implementation
   - Extracts template path from HTML comments (`<!-- path -->`)
   - Provides Theme(), Path(), Name(), and Content() methods

3. **Store (`store.go`)**: Template store abstraction
   - `StoreFS` implementation using Go's `fs.FS`
   - Organizes templates by theme directories
   - Template lookup with parent theme fallback support

4. **Functions (`funcs.go`)**: Rich template function library
   - Arithmetic operations (add, sub, mul, div)
   - String manipulation (camelCase, snakeCase, trim, etc.)
   - Type conversions (to_int, to_string, to_time, etc.)
   - Slice operations (first, last, append, reverse, etc.)
   - Map operations (dict, keys, values, get, set, etc.)
   - Time/date formatting with location support
   - Data encoding (JSON, XML, YAML with pretty options)

### Key Features

- **Template Inheritance**: Templates can extend other templates using `{{template "name"}}` or `{{block "name"}}`
- **Theme Fallback**: Child themes automatically fall back to parent themes for missing templates
- **Path Resolution**: Template paths extracted from HTML comments for inheritance chains
- **Caching**: Built-in template caching with debug mode bypass
- **Rich Function Library**: Comprehensive template functions for common operations

## Development Commands

### Build and Test
```bash
go build ./...
go test ./...
go test -v ./...         # Verbose tests
go test -race ./...       # Race detection
```

### Module Management
```bash
go mod tidy              # Clean dependencies
go mod download          # Download dependencies
go mod verify            # Verify dependencies
```

### Code Quality
```bash
go fmt ./...             # Format code
go vet ./...             # Static analysis
golangci-lint run -v --timeout=5m --build-tags=race --output.code-climate.path gl-code-quality-report.json # Lint code
```

## Implementation Details

### Template Path Convention
Templates use HTML comments to define inheritance paths:
```html
<!-- layouts/base -->
{{define "content"}}...{{end}}
```

### Theme Hierarchy
Create parent-child relationships:
```go
parent := NewTheme("default", store)
child := NewTheme("custom", store)
child.SetParent(parent)
```

### Function Map Integration
Custom functions are automatically propagated to parent themes when set.

### Template Resolution
The system resolves templates by:
1. Looking in current theme's store
2. Falling back to parent theme if not found
3. Parsing template dependencies recursively
4. Building complete Go template with all dependencies

## Error Handling

- `ErrTemplateNotFound`: Primary error for missing templates
- Errors include theme and template name for debugging
- Store errors wrap filesystem errors with context

## Performance Considerations

- Templates cached in `sync.Map` for concurrent access
- Debug mode bypasses cache for development
- Template dependencies resolved once per execution
- Parent theme lookups follow inheritance chain

## Testing Strategy

- Comprehensive unit tests with table-driven test patterns using `testify`
- Mock implementations for store and external dependencies using `testify`