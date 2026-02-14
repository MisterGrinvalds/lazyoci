package pull

import (
	"fmt"
	"io"
	"sync"

	"github.com/schollz/progressbar/v3"
)

// ProgressTracker manages progress bars for layer downloads.
type ProgressTracker struct {
	mu    sync.Mutex
	bars  map[string]*progressbar.ProgressBar
	quiet bool
}

// NewProgressTracker creates a new progress tracker.
// If quiet is true, no progress output is shown.
func NewProgressTracker(quiet bool) *ProgressTracker {
	return &ProgressTracker{
		bars:  make(map[string]*progressbar.ProgressBar),
		quiet: quiet,
	}
}

// shortDigest returns a shortened digest for display (first 12 chars after sha256:).
func shortDigest(digest string) string {
	if len(digest) > 19 {
		return digest[:19]
	}
	return digest
}

// StartLayer creates a new progress bar for a layer download.
func (t *ProgressTracker) StartLayer(digest string, size int64) {
	if t.quiet {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// Create progress bar with schollz/progressbar
	bar := progressbar.NewOptions64(
		size,
		progressbar.OptionSetDescription(fmt.Sprintf("Pulling %s", shortDigest(digest))),
		progressbar.OptionSetWidth(30),
		progressbar.OptionShowBytes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionOnCompletion(func() {
			fmt.Println()
		}),
	)

	t.bars[digest] = bar
}

// UpdateLayer updates the progress for a layer.
func (t *ProgressTracker) UpdateLayer(digest string, bytesRead int64) {
	if t.quiet {
		return
	}

	t.mu.Lock()
	bar, ok := t.bars[digest]
	t.mu.Unlock()

	if ok {
		bar.Set64(bytesRead)
	}
}

// FinishLayer marks a layer as complete.
func (t *ProgressTracker) FinishLayer(digest string) {
	if t.quiet {
		return
	}

	t.mu.Lock()
	bar, ok := t.bars[digest]
	t.mu.Unlock()

	if ok {
		bar.Finish()
	}
}

// Finish completes all progress bars.
func (t *ProgressTracker) Finish() {
	if t.quiet {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	for _, bar := range t.bars {
		bar.Finish()
	}
}

// trackingReader wraps an io.Reader to track bytes read for progress.
type trackingReader struct {
	reader  io.Reader
	digest  string
	tracker *ProgressTracker
	total   int64
	current int64
}

func (t *trackingReader) Read(p []byte) (int, error) {
	n, err := t.reader.Read(p)
	if n > 0 {
		t.current += int64(n)
		t.tracker.UpdateLayer(t.digest, t.current)
	}
	return n, err
}

// NewTrackingReader creates a reader that reports progress to the tracker.
func NewTrackingReader(r io.Reader, digest string, size int64, tracker *ProgressTracker) io.Reader {
	tracker.StartLayer(digest, size)
	return &trackingReader{
		reader:  r,
		digest:  digest,
		tracker: tracker,
		total:   size,
	}
}
