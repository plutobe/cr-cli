package git

import (
	"testing"
)

func TestParseChangeset(t *testing.T) {
	diff := `diff --git a/main.go b/main.go
index abc1234..def5678 100644
--- a/main.go
+++ b/main.go
@@ -1,3 +1,4 @@
 package main

-func main() {
+func main() { // changed
+	println("hello")
 }
diff --git a/binary.png b/binary.png
Binary files differ
`
	files := ParseChangeset(diff)
	if len(files) != 1 {
		t.Fatalf("expected 1 file (skip binary), got %d", len(files))
	}
	if files[0].Filename != "main.go" {
		t.Errorf("filename = %q, want %q", files[0].Filename, "main.go")
	}
	if files[0].Language != "go" {
		t.Errorf("language = %q, want %q", files[0].Language, "go")
	}
}

func TestParseChangeset_FilterByIgnorePatterns(t *testing.T) {
	diff := `diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -1,1 +1,2 @@
 package main
+// change
diff --git a/vendor/lib.go b/vendor/lib.go
--- a/vendor/lib.go
+++ b/vendor/lib.go
@@ -1,1 +1,2 @@
 package vendor
+// change
diff --git a/node_modules/pkg/index.js b/node_modules/pkg/index.js
--- a/node_modules/pkg/index.js
+++ b/node_modules/pkg/index.js
@@ -1,1 +1,2 @@
 console.log("hi")
+// change
`
	files := ParseChangeset(diff, WithIgnorePatterns([]string{"vendor/**", "node_modules/**"}))
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if files[0].Filename != "main.go" {
		t.Errorf("filename = %q, want %q", files[0].Filename, "main.go")
	}
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"main.go", "go"},
		{"app.py", "python"},
		{"index.js", "javascript"},
		{"app.ts", "typescript"},
		{"Main.java", "java"},
		{"main.c", "c"},
		{"main.cpp", "cpp"},
		{"lib.rs", "rust"},
		{"unknown.xyz", ""},
	}
	for _, tt := range tests {
		got := detectLanguage(tt.filename)
		if got != tt.want {
			t.Errorf("detectLanguage(%q) = %q, want %q", tt.filename, got, tt.want)
		}
	}
}
