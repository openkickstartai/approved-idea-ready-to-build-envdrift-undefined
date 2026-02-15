package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var skipDirs = map[string]bool{
	".git": true, "vendor": true, "node_modules": true, "__pycache__": true,
}

var scanExts = map[string]bool{
	".go": true, ".py": true, ".js": true, ".ts": true, ".jsx": true, ".tsx": true,
}

func main() {
	dir := flag.String("dir", ".", "root directory to scan")
	envFile := flag.String("env", ".env.example", "env definition file to compare against")
	format := flag.String("format", "text", "output format: text or json")
	flag.Parse()

	codeVars := map[string][]Location{}
	err := filepath.Walk(*dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if skipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if !scanExts[strings.ToLower(filepath.Ext(path))] {
			return nil
		}
		vars, scanErr := ScanFile(path)
		if scanErr != nil {
			fmt.Fprintf(os.Stderr, "warn: %s: %v\n", path, scanErr)
			return nil
		}
		for k, v := range vars {
			codeVars[k] = append(codeVars[k], v...)
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	defined, err := ParseEnvFile(*envFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading %s: %v\n", *envFile, err)
		os.Exit(2)
	}

	report := ComputeDrift(codeVars, defined)
	if *format == "json" {
		fmt.Println(report.JSON())
	} else {
		fmt.Print(report.Text())
	}
	if len(report.Missing) > 0 {
		os.Exit(1)
	}
}
