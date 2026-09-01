package utils

import (
	"os"
	"runtime"
	"strings"
	"testing"
)

func TestCallerPrettyfier_StripsProjectRoot(t *testing.T) {
	frame := &runtime.Frame{
		File:     projectRoot + "dataproc/foo.go",
		Line:     42,
		Function: "github.com/Kaszanas/SC2InfoExtractorGo/dataproc.Foo",
	}

	function, file := callerPrettyfier(frame)

	wantFile := "dataproc/foo.go:42"
	if file != wantFile {
		t.Errorf("callerPrettyfier() file = %q, want %q", file, wantFile)
	}
	if function != frame.Function {
		t.Errorf("callerPrettyfier() function = %q, want %q (unchanged)", function, frame.Function)
	}
}

func TestCallerPrettyfier_PassesThroughUnknownPrefix(t *testing.T) {
	frame := &runtime.Frame{
		File: "/usr/local/go/src/runtime/proc.go",
		Line: 7,
	}

	_, file := callerPrettyfier(frame)

	wantFile := "/usr/local/go/src/runtime/proc.go:7"
	if file != wantFile {
		t.Errorf("callerPrettyfier() file = %q, want %q", file, wantFile)
	}
}

func TestSetLogging_NoAbsoluteBuildPathLeaksIntoLogFile(t *testing.T) {
	tempDir := t.TempDir() + string(os.PathSeparator)

	logFile, ok := SetLogging(tempDir, 4)
	if !ok {
		t.Fatal("SetLogging() returned ok = false")
	}
	defer func() {
		if err := logFile.Close(); err != nil {
			t.Errorf("failed to close log file: %v", err)
		}
	}()

	contents, err := os.ReadFile(logFile.Name())
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	if projectRoot != "" && strings.Contains(string(contents), projectRoot) {
		t.Errorf("log file contains the absolute build path prefix %q:\n%s", projectRoot, contents)
	}
}
