// apiforge CLI:
//   apiforge lint  <spec.json>
//   apiforge diff  <old.json> <new.json>
//   apiforge mock  <spec.json> [--addr 127.0.0.1:8080]

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/SAY-5/apiforge/diff"
	"github.com/SAY-5/apiforge/lint"
	"github.com/SAY-5/apiforge/mock"
	"github.com/SAY-5/apiforge/spec"
)

func usage() {
	fmt.Fprintln(os.Stderr, "usage: apiforge {lint|diff|mock} ...")
}

func loadSpec(path string) *spec.Spec {
	raw, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	s, err := spec.Parse(raw)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	return s
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "lint":
		if len(os.Args) < 3 {
			usage()
			os.Exit(2)
		}
		s := loadSpec(os.Args[2])
		findings := lint.Run(s, lint.DefaultRules())
		_ = json.NewEncoder(os.Stdout).Encode(findings)
		errs := 0
		for _, f := range findings {
			if f.Severity == lint.SeverityError {
				errs++
			}
		}
		if errs > 0 {
			os.Exit(1)
		}
	case "diff":
		if len(os.Args) < 4 {
			usage()
			os.Exit(2)
		}
		old := loadSpec(os.Args[2])
		new := loadSpec(os.Args[3])
		changes := diff.Compare(old, new)
		_ = json.NewEncoder(os.Stdout).Encode(changes)
		if diff.HasBreaking(changes) {
			os.Exit(1)
		}
	case "mock":
		if len(os.Args) < 3 {
			usage()
			os.Exit(2)
		}
		s := loadSpec(os.Args[2])
		addr := "127.0.0.1:8080"
		if len(os.Args) >= 5 && os.Args[3] == "--addr" {
			addr = os.Args[4]
		}
		fmt.Fprintln(os.Stderr, "mock server listening on", addr)
		_ = http.ListenAndServe(addr, mock.NewServer(s))
	default:
		usage()
		os.Exit(2)
	}
}
