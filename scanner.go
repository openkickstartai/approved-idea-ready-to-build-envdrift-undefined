package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

type Location struct {
	File string `json:"file"`
	Line int    `json:"line"`
}

type DriftItem struct {
	Name      string     `json:"name"`
	Locations []Location `json:"locations,omitempty"`
}

type Report struct {
	Missing []DriftItem `json:"missing"`
	Unused  []DriftItem `json:"unused"`
}

var patterns = []*regexp.Regexp{
	regexp.MustCompile(`os\.Getenv\("([A-Z_][A-Z0-9_]*)"\)`),
	regexp.MustCompile(`os\.LookupEnv\("([A-Z_][A-Z0-9_]*)"\)`),
	regexp.MustCompile(`process\.env\.([A-Z_][A-Z0-9_]*)\b`),
	regexp.MustCompile(`process\.env\[["']([A-Z_][A-Z0-9_]*)["']\]`),
	regexp.MustCompile(`os\.environ\[["']([A-Z_][A-Z0-9_]*)["']\]`),
	regexp.MustCompile(`os\.environ\.get\(["']([A-Z_][A-Z0-9_]*)["']`),
	regexp.MustCompile(`os\.getenv\(["']([A-Z_][A-Z0-9_]*)["']\)`),
}

func ScanFile(path string) (map[string][]Location, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	result := map[string][]Location{}
	s := bufio.NewScanner(f)
	lineNo := 0
	for s.Scan() {
		lineNo++
		text := s.Text()
		for _, p := range patterns {
			for _, m := range p.FindAllStringSubmatch(text, -1) {
				result[m[1]] = append(result[m[1]], Location{File: path, Line: lineNo})
			}
		}
	}
	return result, s.Err()
}

func ParseEnvFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	result := map[string]string{}
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(parts[0])
		val := ""
		if len(parts) == 2 {
			val = strings.TrimSpace(parts[1])
		}
		result[key] = val
	}
	return result, s.Err()
}

func ComputeDrift(codeVars map[string][]Location, defined map[string]string) Report {
	var r Report
	for name, locs := range codeVars {
		if _, ok := defined[name]; !ok {
			r.Missing = append(r.Missing, DriftItem{Name: name, Locations: locs})
		}
	}
	for name := range defined {
		if _, ok := codeVars[name]; !ok {
			r.Unused = append(r.Unused, DriftItem{Name: name})
		}
	}
	sort.Slice(r.Missing, func(i, j int) bool { return r.Missing[i].Name < r.Missing[j].Name })
	sort.Slice(r.Unused, func(i, j int) bool { return r.Unused[i].Name < r.Unused[j].Name })
	return r
}

func (r Report) JSON() string {
	b, _ := json.MarshalIndent(r, "", "  ")
	return string(b)
}

func (r Report) Text() string {
	var sb strings.Builder
	if len(r.Missing) > 0 {
		fmt.Fprintf(&sb, "MISSING — referenced in code but not in env file (%d):\n", len(r.Missing))
		for _, item := range r.Missing {
			fmt.Fprintf(&sb, "  ✗ %s\n", item.Name)
			for _, loc := range item.Locations {
				fmt.Fprintf(&sb, "      %s:%d\n", loc.File, loc.Line)
			}
		}
	}
	if len(r.Unused) > 0 {
		fmt.Fprintf(&sb, "UNUSED — defined in env file but not in code (%d):\n", len(r.Unused))
		for _, item := range r.Unused {
			fmt.Fprintf(&sb, "  ? %s\n", item.Name)
		}
	}
	if len(r.Missing) == 0 && len(r.Unused) == 0 {
		sb.WriteString("✓ No drift detected.\n")
	}
	return sb.String()
}
