package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/schollz/progressbar/v3"
)

// NewProgressBar creates a new progress bar standardized for the project.
func NewProgressBar(
	lenghtOfProgressbar int,
	progressBarDescription string,
) *progressbar.ProgressBar {

	progressBar := progressbar.NewOptions(
		lenghtOfProgressbar,
		progressbar.
			OptionSetDescription(
				progressBarDescription,
			),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionSetWidth(10),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n")
		}),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetRenderBlankState(true),
	)

	return progressBar
}
