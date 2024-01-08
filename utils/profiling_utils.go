package utils

import (
	"os"
	"runtime/pprof"

	log "github.com/sirupsen/logrus"
)

// setProfiling sets up pprof profiling to a supplied filepath.
func SetProfiling(profilingPath string) (*os.File, bool) {

	performCPUProfilingPath := profilingPath

	// Creating profiler file:
	profilerFile, err := os.Create(performCPUProfilingPath)
	if err != nil {
		log.WithField("error", err).Error("Could not create a profiling file. Exiting program.")
		return profilerFile, false
	}
	// Starting profiling:
	err = pprof.StartCPUProfile(profilerFile)
	if err != nil {
		log.WithField("error", err).Error("Could not start profiling. Exiting program.")
		return profilerFile, false
	}

	return profilerFile, true
}
