package utils

import (
	"os"
	"testing"

	"github.com/Kaszanas/SC2InfoExtractorGo/settings"
)

func TestSetProfilingEmpty(t *testing.T) {

	_, profilingSetOk := SetProfiling("")

	if profilingSetOk {
		t.Fatalf("Test Failed! setProfiling returned true on an empty string!.")
	}
}

func TestSetProfiling(t *testing.T) {

	testProfilerPath, err := settings.GetProfilerReportPath()
	if err != nil {
		t.Fatalf("Test Failed! Couldn't get the test profiler path.")
	}

	profilerFile, profilingSetOk := SetProfiling(testProfilerPath)

	if !profilingSetOk {
		t.Fatalf("Test Failed! setProfiling returned false on a valid path.")
	}

	err = profilerFile.Close()
	if err != nil {
		t.Fatalf("Test Failed! Couldn't close the profiling file.")
	}

	err = os.Remove(testProfilerPath)
	if err != nil {
		t.Fatalf("Test Failed! Cannot delete profiling file.")
	}

}
