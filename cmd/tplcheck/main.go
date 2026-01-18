// Package main provides a CLI tool to compile and validate TPL templates.
// It walks a directory tree, finds directories containing _properties.json,
// and compiles all .tpl files in those directories, reporting any errors.
package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/KarpelesLab/tpl"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <directory>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Compile and validate TPL templates in a directory tree.\n")
		fmt.Fprintf(os.Stderr, "Finds directories containing _properties.json and compiles all .tpl files.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	verbose := flag.Bool("v", false, "verbose output")
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	rootDir := flag.Arg(0)

	// Verify the directory exists
	info, err := os.Stat(rootDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: %s is not a directory\n", rootDir)
		os.Exit(1)
	}

	// Find all directories containing _properties.json
	templateDirs, err := findTemplateDirs(rootDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning directory: %v\n", err)
		os.Exit(1)
	}

	if len(templateDirs) == 0 {
		fmt.Println("No template directories found (no _properties.json files)")
		os.Exit(0)
	}

	if *verbose {
		fmt.Printf("Found %d template directories\n", len(templateDirs))
	}

	// Compile templates in each directory
	ctx := context.Background()
	hasErrors := false
	successCount := 0

	for _, dir := range templateDirs {
		relDir, _ := filepath.Rel(rootDir, dir)
		if relDir == "" || relDir == "." {
			relDir = filepath.Base(dir)
		}

		errs := compileTemplates(ctx, dir)
		if len(errs) > 0 {
			hasErrors = true
			fmt.Printf("FAIL %s\n", relDir)
			for _, e := range errs {
				fmt.Printf("  %v\n", e)
			}
		} else {
			successCount++
			if *verbose {
				fmt.Printf("OK   %s\n", relDir)
			}
		}
	}

	fmt.Printf("\nResults: %d/%d directories compiled successfully\n", successCount, len(templateDirs))

	if hasErrors {
		os.Exit(1)
	}
}

// findTemplateDirs finds all directories containing _properties.json
func findTemplateDirs(root string) ([]string, error) {
	var dirs []string

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && d.Name() == "_properties.json" {
			dirs = append(dirs, filepath.Dir(path))
		}

		return nil
	})

	return dirs, err
}

// compileTemplates compiles all .tpl files in a directory and returns any errors
func compileTemplates(ctx context.Context, dir string) []error {
	var errs []error

	// Find all .tpl files in the directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		return []error{fmt.Errorf("failed to read directory: %w", err)}
	}

	// Create a new template engine
	engine := tpl.New()

	// Load each .tpl file
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tpl") {
			continue
		}

		tplName := strings.TrimSuffix(entry.Name(), ".tpl")
		tplPath := filepath.Join(dir, entry.Name())

		content, err := os.ReadFile(tplPath)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: failed to read: %w", entry.Name(), err))
			continue
		}

		engine.Raw.TemplateData[tplName] = string(content)
	}

	if len(engine.Raw.TemplateData) == 0 {
		return nil // No templates to compile
	}

	// Compile all templates
	if err := engine.Compile(ctx); err != nil {
		errs = append(errs, err)
	}

	return errs
}
