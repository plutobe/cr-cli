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
	ignorePatterns []string
}

// WithIgnorePatterns filters out files matching any of the glob patterns.
func WithIgnorePatterns(patterns []string) ParseOption {
	return func(o *parseOptions) {
		o.ignorePatterns = patterns
	}
}

// ParseChangeset parses unified diff output into individual file changes.
// Binary files are skipped. Options can filter by ignore patterns.
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
	for _, pattern := range opts.ignorePatterns {
		if matchIgnorePattern(pattern, f.Filename) {
			return files
		}
	}
	return append(files, f)
}

func matchIgnorePattern(pattern, filename string) bool {
	// Directory pattern: "vendor/**" matches "vendor/anything"
	if strings.HasSuffix(pattern, "/**") {
		dir := strings.TrimSuffix(pattern, "/**")
		return strings.HasPrefix(filename, dir+"/")
	}
	// Glob match against basename
	if matched, _ := filepath.Match(pattern, filepath.Base(filename)); matched {
		return true
	}
	// Glob match against full path
	if matched, _ := filepath.Match(pattern, filename); matched {
		return true
	}
	return false
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
		".html": "html", ".htm": "html", ".css": "css",
		".scss": "css", ".less": "css", ".vue": "javascript",
		".jsx": "javascript", ".tsx": "typescript",
	}
	if lang, ok := langMap[ext]; ok {
		return lang
	}
	return ""
}
