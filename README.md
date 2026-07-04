# go-sarif

[![Go Reference](https://pkg.go.dev/badge/github.com/richardwooding/go-sarif.svg)](https://pkg.go.dev/github.com/richardwooding/go-sarif)

**Website:** [richardwooding.github.io/go-sarif](https://richardwooding.github.io/go-sarif/)

A **minimal, zero-dependency** SARIF 2.1.0 emitter for Go static-analysis tools
— map findings to results, name your tool, write the document that GitHub Code
Scanning / GitLab and other CI systems ingest.

Standard library only. One run, one driver, line-level result regions — the
subset most analyzers need to surface findings. For richer SARIF (multiple runs,
help URIs, fingerprints, suppressions) use a fuller library such as
[`owenrumney/go-sarif`](https://github.com/owenrumney/go-sarif).

## Install

```sh
go get github.com/richardwooding/go-sarif
```

## Usage

```go
import (
	"os"

	sarif "github.com/richardwooding/go-sarif"
)

func main() {
	_ = sarif.Write(os.Stdout,
		sarif.Tool{
			Name:           "mytool",
			Version:        "1.2.3",
			InformationURI: "https://example.com/mytool",
		},
		[]sarif.Rule{
			{ID: "complexity", Name: "CyclomaticComplexity", Description: "Functions over the complexity ceiling"},
		},
		[]sarif.Result{
			{RuleID: "complexity", Level: "error", Message: "F is too complex", URI: "pkg/a.go", StartLine: 10, EndLine: 42},
			{RuleID: "complexity", Message: "file-level finding", URI: "pkg/b.go"}, // no line → file-level result
		},
	)
}
```

```go
type Tool struct {
	Name           string // driver name surfaced in Code Scanning (required)
	Version        string // optional; omitted when empty
	InformationURI string // optional; omitted when empty
}

type Result struct {
	RuleID    string
	Level     string // "error" | "warning" | "note"; defaults to "warning"
	Message   string
	URI       string // file path, relative to the repo root
	StartLine int    // <= 0 → file-level result (no region)
	EndLine   int
}

type Rule struct{ ID, Name, Description string }

func Write(w io.Writer, tool Tool, rules []Rule, results []Result) error
```

Notes:

- Artifact URIs are emitted with forward slashes (RFC 3986), so paths are
  portable across Windows and POSIX.
- A `Result` with `StartLine <= 0` produces a file-level result (no `region`) —
  for findings without a single line (dead code, unused exports, …).
- An empty `results` slice still writes a valid run.

Extracted from [file-search-on](https://github.com/richardwooding/file-search-on).

## License

MIT — see [LICENSE](LICENSE).
