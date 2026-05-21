package git

import (
	"path/filepath"
	"strings"
)

// FileChange represents a single file's diff and detected language.
type FileChange struct {
	Filename string
	Language string
	Diff     string
}

// ParseOption configures changeset parsing.
type ParseOption func(*parseOptions)

type parseOptions struct {
	languages []string
}

// WithLanguages filters results to only include files matching the given languages.
func WithLanguages(langs []string) ParseOption {
	return func(o *parseOptions) {
		o.languages = langs
	}
}

// ParseChangeset parses unified diff output into individual file changes.
// Binary files are skipped. Options can filter by language.
func ParseChangeset(diff string, opts ...ParseOption) []FileChange {
	options := &parseOptions{}
	for _, o := range opts {
		o(options)
	}

	var files []FileChange
	var current *FileChange
	var lines []string

	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "diff --git ") {
			if current != nil {
				current.Diff = strings.Join(lines, "\n")
				files = appendIfAllowed(files, *current, options)
			}
			parts := strings.Split(line, " b/")
			filename := parts[len(parts)-1]
			current = &FileChange{
				Filename: filename,
				Language: detectLanguage(filename),
			}
			lines = []string{}
			continue
		}
		if current != nil {
			if strings.HasPrefix(line, "Binary files") {
				current = nil
				lines = nil
				continue
			}
			lines = append(lines, line)
		}
	}
	if current != nil {
		current.Diff = strings.Join(lines, "\n")
		files = appendIfAllowed(files, *current, options)
	}
	return files
}

func appendIfAllowed(files []FileChange, f FileChange, opts *parseOptions) []FileChange {
	if len(opts.languages) > 0 {
		if f.Language == "" {
			return files
		}
		found := false
		for _, l := range opts.languages {
			if l == f.Language {
				found = true
				break
			}
		}
		if !found {
			return files
		}
	}
	return append(files, f)
}

func detectLanguage(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	langMap := map[string]string{
		".go": "go", ".py": "python", ".js": "javascript",
		".ts": "typescript", ".java": "java", ".c": "c",
		".cpp": "cpp", ".cc": "cpp", ".cxx": "cpp",
		".rs": "rust", ".rb": "ruby", ".php": "php",
		".swift": "swift", ".kt": "kotlin", ".scala": "scala",
		".sh": "shell", ".bash": "shell", ".sql": "sql",
	}
	if lang, ok := langMap[ext]; ok {
		return lang
	}
	return ""
}
