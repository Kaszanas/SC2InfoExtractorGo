package cmd

import (
	"bytes"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Kaszanas/SC2InfoExtractorGo/datastruct"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils"
)

// fakeRunner records which runner method was called and with what configuration.
type fakeRunner struct {
	called string
	flags  utils.CLIFlags
}

func (f *fakeRunner) Process(flags utils.CLIFlags) error {
	f.called, f.flags = "process", flags
	return nil
}

func (f *fakeRunner) DownloadDeps(flags utils.CLIFlags) error {
	f.called, f.flags = "download_deps", flags
	return nil
}

func execute(t *testing.T, args ...string) (*fakeRunner, error) {
	t.Helper()
	runner := &fakeRunner{}
	rootCmd := NewRootCmd(runner)
	rootCmd.SetArgs(args)
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})
	return runner, rootCmd.Execute()
}

func mustAbs(t *testing.T, path string) string {
	t.Helper()
	absPath, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("filepath.Abs(%q): %v", path, err)
	}
	return absPath
}

func TestCommandsBuildExpectedFlags(t *testing.T) {
	defaults := func() utils.CLIFlags {
		return utils.CLIFlags{
			InputDirectory:      mustAbs(t, "./replays/input"),
			OutputDirectory:     mustAbs(t, "./replays/output"),
			DependencyDirectory: mustAbs(t, "./dependencies/"),
			NumberOfThreads:     runtime.NumCPU(),
			FilterGameMode:      0b11111111,
			LogFlags: utils.LogFlags{
				LogLevelValue: datastruct.Warn,
				LogPath:       "./logs/",
			},
		}
	}

	testCases := []struct {
		name       string
		args       []string
		wantCalled string
		want       func() utils.CLIFlags
	}{
		{
			name:       "json_zip defaults",
			args:       []string{"process", "json_zip"},
			wantCalled: "process",
			want: func() utils.CLIFlags {
				flags := defaults()
				flags.NumberOfPackages = 1
				return flags
			},
		},
		{
			name: "json_zip with paths and packages",
			args: []string{"process", "json_zip",
				"--input", "in", "--output", "out", "--dependency_directory", "deps",
				"--number_of_packages", "3", "--max_procs", "2"},
			wantCalled: "process",
			want: func() utils.CLIFlags {
				flags := defaults()
				flags.InputDirectory = mustAbs(t, "in")
				flags.OutputDirectory = mustAbs(t, "out")
				flags.DependencyDirectory = mustAbs(t, "deps")
				flags.NumberOfPackages = 3
				flags.NumberOfThreads = 2
				return flags
			},
		},
		{
			name:       "json writes one file per replay",
			args:       []string{"process", "json"},
			wantCalled: "process",
			want:       defaults,
		},
		{
			name:       "single_json",
			args:       []string{"process", "single_json"},
			wantCalled: "process",
			want: func() utils.CLIFlags {
				flags := defaults()
				flags.SingleJsonOutput = true
				flags.NumberOfPackages = 1
				return flags
			},
		},
		{
			name: "processing switches",
			args: []string{"process", "json",
				"--skip_dependency_download", "--perform_integrity_checks",
				"--perform_validity_checks", "--perform_cleanup",
				"--perform_player_anonymization", "--perform_chat_anonymization",
				"--perform_filtering", "--game_mode_filter", "0b101",
				"--with_cpu_profiler", "cpu.prof"},
			wantCalled: "process",
			want: func() utils.CLIFlags {
				flags := defaults()
				flags.SkipDependencyDownload = true
				flags.PerformIntegrityCheck = true
				flags.PerformValidityCheck = true
				flags.PerformCleanup = true
				flags.PerformPlayerAnonymization = true
				flags.PerformChatAnonymization = true
				flags.PerformFiltering = true
				flags.FilterGameMode = 0b101
				flags.CPUProfilingPath = "cpu.prof"
				return flags
			},
		},
		{
			name:       "snake_case log flags",
			args:       []string{"process", "json", "--log_level", "5", "--log_dir", "custom_logs/"},
			wantCalled: "process",
			want: func() utils.CLIFlags {
				flags := defaults()
				flags.LogFlags = utils.LogFlags{LogLevelValue: datastruct.Debug, LogPath: "custom_logs/"}
				return flags
			},
		},
		{
			name:       "kebab-case spelling is accepted",
			args:       []string{"process", "json", "--log-level", "5", "--log-dir", "custom_logs/"},
			wantCalled: "process",
			want: func() utils.CLIFlags {
				flags := defaults()
				flags.LogFlags = utils.LogFlags{LogLevelValue: datastruct.Debug, LogPath: "custom_logs/"}
				return flags
			},
		},
		{
			name:       "download_deps",
			args:       []string{"download_deps", "--input", "in"},
			wantCalled: "download_deps",
			want: func() utils.CLIFlags {
				flags := defaults()
				flags.InputDirectory = mustAbs(t, "in")
				// download_deps has no processing flags:
				flags.OutputDirectory = ""
				flags.FilterGameMode = 0
				flags.OnlyDependencyDownload = true
				return flags
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			runner, err := execute(t, testCase.args...)
			if err != nil {
				t.Fatalf("Execute(%v) returned error: %v", testCase.args, err)
			}
			if runner.called != testCase.wantCalled {
				t.Fatalf("called %q, want %q", runner.called, testCase.wantCalled)
			}
			if want := testCase.want(); runner.flags != want {
				t.Fatalf("flags mismatch\n got: %+v\nwant: %+v", runner.flags, want)
			}
		})
	}
}

func TestCommandsRejectInvalidInvocations(t *testing.T) {
	testCases := []struct {
		name string
		args []string
	}{
		{"zero packages", []string{"process", "json_zip", "--number_of_packages", "0"}},
		{"old single-dash flag", []string{"-input", "in"}},
		{"flags without a command", []string{"--input", "in"}},
		{"misspelled command", []string{"proces", "json"}},
		{"process without a format", []string{"process", "--output", "out"}},
		{"unknown format", []string{"process", "xml"}},
		{"removed flag", []string{"process", "json_zip", "--single_json_output"}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			runner, err := execute(t, testCase.args...)
			if err == nil {
				t.Fatalf("Execute(%v) succeeded, want an error", testCase.args)
			}
			if runner.called != "" {
				t.Fatalf("runner %q was called for an invalid invocation", runner.called)
			}
		})
	}
}

func TestGroupCommandsShowHelp(t *testing.T) {
	for _, args := range [][]string{{}, {"process"}} {
		runner, err := execute(t, args...)
		if err != nil {
			t.Fatalf("Execute(%v) returned error: %v", args, err)
		}
		if runner.called != "" {
			t.Fatalf("Execute(%v) called runner %q, want help only", args, runner.called)
		}
	}
}
