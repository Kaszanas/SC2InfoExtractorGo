package main

import (
	"os"
	"testing"
)

func TestSetProfilingEmpty(t *testing.T) {

	_, profilingSetOk := setProfiling("")

	if profilingSetOk {
		t.Fatalf("Test Failed! setProfiling returned true on an empty string!.")
	}
}

func TestSetProfiling(t *testing.T) {

	profilerPath := "./test_files/test_profiler.txt"

	profilerFile, profilingSetOk := setProfiling(profilerPath)

	if !profilingSetOk {
		t.Fatalf("Test Failed! setProfiling returned false on a valid path.")
	}

	err := profilerFile.Close()
	if err != nil {
		t.Fatalf("Test Failed! Couldn't close the profiling file.")
	}

	err = os.Remove(profilerPath)
	if err != nil {
		t.Fatalf("Test Failed! Cannot delete profiling file.")
	}

}
