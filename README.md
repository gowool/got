# GO Theme

[![Go Reference](https://pkg.go.dev/badge/github.com/gowool/got.svg)](https://pkg.go.dev/github.com/gowool/got)
[![Go Report Card](https://goreportcard.com/badge/github.com/gowool/got)](https://goreportcard.com/report/github.com/gowool/got)
[![codecov](https://codecov.io/github/gowool/got/graph/badge.svg?token=U23BO6XII4)](https://codecov.io/github/gowool/got)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](https://github.com/gowool/got/blob/main/LICENSE)

A Go template theme management library that provides a flexible system for organizing and rendering HTML templates with support for template inheritance, themes, and parent-child relationships.

## Installation

```bash
go get github.com/gowool/got
```

## Quick Start

```go
package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	
    "github.com/gowool/got"
)

func main() {
	// Create storage from filesystem
	storage := got.NewStorageFS(os.DirFS("themes"))

	// Create theme
	theme := got.NewTheme("default", storage)
	
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)

		// Render template
		if err := theme.Write(r.Context(), w, "page/index.gohtml", map[string]any{
			"Title": "Hello World",
			"Content": "Welcome to GO Theme",
		}); err != nil {
			log.Println(err.Error())
		}
	})
	
	if err := http.ListenAndServe(":8080", nil); err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}
}
```

## Template Structure

Templates use HTML comments to define inheritance paths:

```html
<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}}</title>
</head>
<body>
    {{block "content" .}}Default content{{end}}
</body>
</html>
```

```html
<!-- layouts/base.gohtml -->

{{define "content"}}
    <h1>{{.Title}}</h1>
    <p>{{.Content}}</p>
{{end}}
```

## Theme Inheritance

Create parent-child theme relationships:

```go
parent := got.NewTheme("default", storage)
child := got.NewTheme("custom", storage)
child.SetParent(parent)

// Child theme will fallback to parent for missing templates
```

## Storage Backends

### Filesystem Storage
```go
storage := got.NewStorageFS(os.DirFS("themes"))
```

### Memory Storage
```go
storage := got.NewStorageMemory()
storage.SetTemplate("theme", "template.html", "content")
```

### Chain Storage
```go
chain := got.NewStorageChain()
chain.Add(memoryStorage)
chain.Add(fsStorage)
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
