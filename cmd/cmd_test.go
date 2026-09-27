package cmd

// These tests check the command line layer only: that each command and flag
// combination turns into the right utils.CLIFlags, and that invalid
// invocations are rejected. No replays are processed here.
//
// How that works: NewRootCmd takes a Runner (the part that does the real
// work). The tests pass a fakeRunner instead of DefaultRunner(), so when a
// command finishes parsing its flags it hands the CLIFlags to the fake, which
// just remembers them. The tests then compare what was remembered with what
// they expected.

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Kaszanas/SC2InfoExtractorGo/datastruct"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils"
)

// fakeRunner records which runner method was called and with what configuration.
// It implements the Runner interface from runner.go without doing any work.
type fakeRunner struct {
	// called is the name of the method that ran ("process", "download_deps"
	// or "process_replay"). It stays empty if no method ran, which is how the
	// tests detect that an invalid invocation was stopped before any work.
	called     string
	flags      utils.CLIFlags
	replayFile string
	// replayJSON and replayErr are returned by ProcessReplay.
	// Tests set them up front to simulate a successful or failed run.
	replayJSON string
	replayErr  error
}

func (f *fakeRunner) Process(flags utils.CLIFlags) error {
	f.called, f.flags = "process", flags
	return nil
}

func (f *fakeRunner) DownloadDeps(flags utils.CLIFlags) error {
	f.called, f.flags = "download_deps", flags
	return nil
}

func (f *fakeRunner) ProcessReplay(flags utils.CLIFlags, replayFile string) (string, error) {
	f.called, f.flags, f.replayFile = "process_replay", flags, replayFile
	return f.replayJSON, f.replayErr
}

// execute runs the CLI with the given arguments, exactly as if they had been
// typed after the program name, and returns the fake runner so the test can
// inspect what the command asked it to do.
func execute(t *testing.T, args ...string) (*fakeRunner, error) {
	t.Helper()
	runner := &fakeRunner{}
	// A fresh command tree per call: cobra stores parsed flag values inside
	// the tree, so reusing one would leak flags from one test into the next.
	rootCmd := NewRootCmd(runner)
	// SetArgs replaces os.Args, so the test controls the command line.
	rootCmd.SetArgs(args)
	// Send help and error text to throwaway buffers to keep test output quiet.
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})
	return runner, rootCmd.Execute()
}

// mustAbs returns the absolute form of a path, the way the commands resolve
// --input, --output and --dependency_directory. Expected values are built
// with it so the tests pass no matter which directory they run from.
func mustAbs(t *testing.T, path string) string {
	t.Helper()
	absPath, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("filepath.Abs(%q): %v", path, err)
	}
	return absPath
}

// TestCommandsBuildExpectedFlags is a table-driven test: each entry is one
// command line, plus the runner method and CLIFlags it should produce.
func TestCommandsBuildExpectedFlags(t *testing.T) {
	// defaults is what `process ...` produces when no flags are given.
	// Each test case starts from it and changes only the fields its flags
	// should affect, so a case documents exactly what its flags do.
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
		name string
		// args is the command line, without the program name.
		args []string
		// wantCalled is the runner method that should run.
		wantCalled string
		// want builds the CLIFlags the runner should receive. It's a function
		// so each case gets its own fresh copy of defaults() to modify.
		want func() utils.CLIFlags
	}{
		{
			// The output format is picked by the subcommand: json_zip means
			// zip packages, so NumberOfPackages must be at least 1.
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
			// Paths given on the command line are turned into absolute paths.
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
			// NumberOfPackages = 0 is how the pipeline knows to write one
			// .json file per replay instead of zip packages.
			name:       "json writes one file per replay",
			args:       []string{"process", "json"},
			wantCalled: "process",
			want:       defaults,
		},
		{
			// single_json keeps one chunk (NumberOfPackages = 1) so the order
			// of replays in all_replays.json is always the input file order.
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
			// Every processing switch set at once, to check each flag lands
			// in the right CLIFlags field. "0b101" also checks that binary
			// notation is accepted for --game_mode_filter.
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
			// The documented snake_case spelling.
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
			// The same flags spelled with dashes must give the same result,
			// thanks to the snakeCaseFlags normalizer in root.go.
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
			// download_deps calls a different runner method and doesn't have
			// the `process` flags, so those fields stay at their zero values.
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
		// t.Run makes each case its own subtest, so a failure names the case.
		t.Run(testCase.name, func(t *testing.T) {
			// 1. Run the command line against the fake runner.
			runner, err := execute(t, testCase.args...)
			if err != nil {
				t.Fatalf("Execute(%v) returned error: %v", testCase.args, err)
			}
			// 2. The right runner method must have been called...
			if runner.called != testCase.wantCalled {
				t.Fatalf("called %q, want %q", runner.called, testCase.wantCalled)
			}
			// 3. ...with exactly the expected configuration. CLIFlags holds
			// only comparable fields, so != compares every field at once.
			if want := testCase.want(); runner.flags != want {
				t.Fatalf("flags mismatch\n got: %+v\nwant: %+v", runner.flags, want)
			}
		})
	}
}

// TestCommandsRejectInvalidInvocations checks that bad command lines fail
// with an error and never reach the runner, so nothing is processed by accident.
func TestCommandsRejectInvalidInvocations(t *testing.T) {
	testCases := []struct {
		name string
		args []string
	}{
		// json_zip needs at least one package.
		{"zero packages", []string{"process", "json_zip", "--number_of_packages", "0"}},
		// The 2.x syntax: one dash and no command. It must fail loudly
		// rather than be silently misread.
		{"old single-dash flag", []string{"-input", "in"}},
		// Flags alone, with no command to say what to do.
		{"flags without a command", []string{"--input", "in"}},
		// A typo in the command name.
		{"misspelled command", []string{"proces", "json"}},
		// `process` needs an output format subcommand.
		{"process without a format", []string{"process", "--output", "out"}},
		{"unknown format", []string{"process", "xml"}},
		// A 2.x flag that was replaced by the single_json subcommand.
		{"removed flag", []string{"process", "json_zip", "--single_json_output"}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			runner, err := execute(t, testCase.args...)
			// The command must return an error (non-zero exit code for users)...
			if err == nil {
				t.Fatalf("Execute(%v) succeeded, want an error", testCase.args)
			}
			// ...and must not have started any work.
			if runner.called != "" {
				t.Fatalf("runner %q was called for an invalid invocation", runner.called)
			}
		})
	}
}

// testReplayJSON is the JSON the fake runner "produces" for process_replay.
// Its content doesn't matter; the tests only check where it ends up.
const testReplayJSON = `{"replay":"data"}`

// executeProcessReplay runs process_replay against a fake runner returning
// testReplayJSON (or runnerErr) and captures the command's stdout.
func executeProcessReplay(t *testing.T, runnerErr error, args ...string) (*fakeRunner, string, error) {
	t.Helper()
	runner := &fakeRunner{replayJSON: testReplayJSON, replayErr: runnerErr}
	rootCmd := NewRootCmd(runner)
	// Unlike execute, keep stdout in a buffer we can read afterwards:
	// process_replay prints the JSON there, so stdout is part of what's tested.
	stdout := &bytes.Buffer{}
	rootCmd.SetArgs(append([]string{"process_replay"}, args...))
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(&bytes.Buffer{})
	// Run first, then read stdout. Written as one return statement, Go would
	// read stdout before the command ran and always see it empty.
	err := rootCmd.Execute()
	return runner, stdout.String(), err
}

// createReplayFile creates an empty replay file in a temporary directory.
// process_replay checks that the file exists before calling the runner, so the
// tests need a real file; its content is never read because the runner is fake.
func createReplayFile(t *testing.T) string {
	t.Helper()
	// t.TempDir() is deleted automatically when the test ends.
	replayFile := filepath.Join(t.TempDir(), "game.SC2Replay")
	if err := os.WriteFile(replayFile, nil, 0644); err != nil {
		t.Fatalf("creating %s: %v", replayFile, err)
	}
	return replayFile
}

// TestProcessReplayToStdout checks the default mode: without --output, the
// replay JSON goes to stdout so downstream apps can pipe it.
func TestProcessReplayToStdout(t *testing.T) {
	replayFile := createReplayFile(t)

	// One snake_case and one kebab-case processing flag, to check that
	// process_replay accepts the same processing flags as `process`.
	runner, stdout, err := executeProcessReplay(t, nil,
		replayFile, "--perform_cleanup", "--perform-integrity-checks")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	// The runner was asked to process exactly the file given as the argument.
	if runner.called != "process_replay" || runner.replayFile != replayFile {
		t.Fatalf("runner called %q with %q, want process_replay with %q",
			runner.called, runner.replayFile, replayFile)
	}
	// The processing flags reached the runner.
	if !runner.flags.PerformCleanup || !runner.flags.PerformIntegrityCheck {
		t.Fatalf("processing flags not applied: %+v", runner.flags)
	}
	// There's no --input for process_replay; the input directory is the
	// replay's own directory (the pipeline uses it for the replaypack name).
	if runner.flags.InputDirectory != filepath.Dir(replayFile) {
		t.Fatalf("InputDirectory = %q, want %q", runner.flags.InputDirectory, filepath.Dir(replayFile))
	}
	// stdout holds only the JSON plus a newline, nothing else, so a
	// downstream app can parse it directly.
	if stdout != testReplayJSON+"\n" {
		t.Fatalf("stdout = %q, want %q", stdout, testReplayJSON+"\n")
	}
}

// TestProcessReplayToFile checks the two ways --output writes to disk.
func TestProcessReplayToFile(t *testing.T) {
	replayFile := createReplayFile(t)
	outputDirectory := t.TempDir()

	testCases := []struct {
		name string
		// output is the value passed to --output.
		output string
		// wantFile is where the JSON should be written.
		wantFile string
	}{
		// A file path in a directory that doesn't exist yet: the command
		// creates the "nested" directory and writes exactly that file.
		{"file path", filepath.Join(outputDirectory, "nested", "out.json"),
			filepath.Join(outputDirectory, "nested", "out.json")},
		// An existing directory: the file is named after the replay
		// (game.SC2Replay -> game.json), like `process json` does.
		{"existing directory", outputDirectory, filepath.Join(outputDirectory, "game.json")},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, stdout, err := executeProcessReplay(t, nil, replayFile, "--output", testCase.output)
			if err != nil {
				t.Fatalf("Execute returned error: %v", err)
			}
			// When writing to a file, nothing may be printed to stdout.
			if stdout != "" {
				t.Fatalf("stdout = %q, want nothing when writing to a file", stdout)
			}
			// The file holds exactly the JSON the runner returned.
			content, err := os.ReadFile(testCase.wantFile)
			if err != nil {
				t.Fatalf("reading output: %v", err)
			}
			if string(content) != testReplayJSON {
				t.Fatalf("output file = %q, want %q", content, testReplayJSON)
			}
		})
	}
}

// TestProcessReplayFailureWritesNothing checks that when processing fails, the
// command reports the error and leaves no partial output behind. A downstream
// app must never read half a result or an empty file as if it were valid.
func TestProcessReplayFailureWritesNothing(t *testing.T) {
	replayFile := createReplayFile(t)
	outputFile := filepath.Join(t.TempDir(), "out.json")

	// Try both output modes: stdout, then --output to a file.
	for _, args := range [][]string{{replayFile}, {replayFile, "--output", outputFile}} {
		// The fake runner fails, as the real one would on e.g. a failed integrity check.
		_, stdout, err := executeProcessReplay(t, errors.New("integrity check failed"), args...)
		// The runner's error must come back out of the command...
		if err == nil {
			t.Fatalf("Execute(%v) succeeded, want the runner error", args)
		}
		// ...with nothing printed to stdout...
		if stdout != "" {
			t.Fatalf("stdout = %q, want nothing on failure", stdout)
		}
		// ...and no output file created.
		if _, err := os.Stat(outputFile); !os.IsNotExist(err) {
			t.Fatalf("output file exists after a failure (stat error: %v)", err)
		}
	}
}

// TestProcessReplayRejectsInvalidInvocations checks that bad process_replay
// command lines fail before any processing starts.
func TestProcessReplayRejectsInvalidInvocations(t *testing.T) {
	replayFile := createReplayFile(t)

	testCases := []struct {
		name string
		args []string
	}{
		// Exactly one replay file must be given.
		{"no replay", nil},
		{"two replays", []string{replayFile, replayFile}},
		// The file must exist...
		{"missing file", []string{filepath.Join(t.TempDir(), "missing.SC2Replay")}},
		// ...and be a file: directories are what `process` is for.
		{"directory", []string{t.TempDir()}},
		// --input exists on every command (it's a root flag) but has no
		// meaning here, so setting it is treated as a mistake.
		{"--input set", []string{replayFile, "--input", "in"}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			runner, _, err := executeProcessReplay(t, nil, testCase.args...)
			if err == nil {
				t.Fatalf("Execute(%v) succeeded, want an error", testCase.args)
			}
			// The runner must not have been called: nothing was processed.
			if runner.called != "" {
				t.Fatalf("runner %q was called for an invalid invocation", runner.called)
			}
		})
	}
}

// TestGroupCommandsShowHelp checks that the commands which only group other
// commands (the root and `process`) print their help when run on their own,
// instead of failing or doing any work.
func TestGroupCommandsShowHelp(t *testing.T) {
	// {} is the program run with no arguments; {"process"} is `process` alone.
	for _, args := range [][]string{{}, {"process"}} {
		runner, err := execute(t, args...)
		// Showing help is a normal outcome, not an error...
		if err != nil {
			t.Fatalf("Execute(%v) returned error: %v", args, err)
		}
		// ...and nothing may be processed.
		if runner.called != "" {
			t.Fatalf("Execute(%v) called runner %q, want help only", args, runner.called)
		}
	}
}
