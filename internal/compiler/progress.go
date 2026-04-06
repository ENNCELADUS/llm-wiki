package compiler

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// Progress tracks and displays real-time compilation progress.
type Progress struct {
	mu        sync.Mutex
	phase     string
	total     int
	done      int
	errors    int
	current   string
	startTime time.Time
	isTTY     bool

	// Accumulated results for summary
	summarized []string
	concepts   []string
	articles   []string
	failures   []string
}

// NewProgress creates a progress tracker.
func NewProgress() *Progress {
	// Detect if stdout is a terminal (for interactive display)
	info, _ := os.Stderr.Stat()
	isTTY := info.Mode()&os.ModeCharDevice != 0

	return &Progress{
		startTime: time.Now(),
		isTTY:     isTTY,
	}
}

// StartPhase begins tracking a new compilation phase.
func (p *Progress) StartPhase(name string, total int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.phase = name
	p.total = total
	p.done = 0
	p.errors = 0
	p.current = ""

	if total > 0 {
		fmt.Fprintf(os.Stderr, "\n⏳ %s (%d items)\n", name, total)
	} else {
		fmt.Fprintf(os.Stderr, "\n⏳ %s\n", name)
	}
}

// ItemStart marks the beginning of processing an item.
func (p *Progress) ItemStart(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.current = name

	if p.isTTY && p.total > 0 {
		// Overwrite current line with progress bar
		bar := p.progressBar()
		fmt.Fprintf(os.Stderr, "\r  %s %s", bar, truncatePath(name, 50))
	}
}

// ItemDone marks successful completion of an item.
func (p *Progress) ItemDone(name string, detail string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.done++

	if p.isTTY {
		// Clear the progress line and print the result
		fmt.Fprintf(os.Stderr, "\r  ✓ %s", truncatePath(name, 60))
		if detail != "" {
			fmt.Fprintf(os.Stderr, " → %s", detail)
		}
		fmt.Fprintln(os.Stderr)
	} else {
		// Non-TTY: simple line output
		fmt.Fprintf(os.Stderr, "  [%d/%d] ✓ %s", p.done, p.total, name)
		if detail != "" {
			fmt.Fprintf(os.Stderr, " → %s", detail)
		}
		fmt.Fprintln(os.Stderr)
	}
}

// ItemError marks a failed item.
func (p *Progress) ItemError(name string, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.done++
	p.errors++
	p.failures = append(p.failures, fmt.Sprintf("%s: %s", name, err))

	if p.isTTY {
		fmt.Fprintf(os.Stderr, "\r  ✗ %s — %s\n", truncatePath(name, 50), err)
	} else {
		fmt.Fprintf(os.Stderr, "  [%d/%d] ✗ %s — %s\n", p.done, p.total, name, err)
	}
}

// ConceptsDiscovered reports concepts found during extraction.
func (p *Progress) ConceptsDiscovered(concepts []string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.concepts = append(p.concepts, concepts...)

	if len(concepts) > 0 {
		preview := concepts
		if len(preview) > 5 {
			preview = preview[:5]
		}
		fmt.Fprintf(os.Stderr, "  💡 %d concepts: %s", len(concepts), strings.Join(preview, ", "))
		if len(concepts) > 5 {
			fmt.Fprintf(os.Stderr, " (+%d more)", len(concepts)-5)
		}
		fmt.Fprintln(os.Stderr)
	}
}

// EndPhase marks phase completion.
func (p *Progress) EndPhase() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.total > 0 {
		fmt.Fprintf(os.Stderr, "  Done: %d/%d", p.done, p.total)
		if p.errors > 0 {
			fmt.Fprintf(os.Stderr, " (%d errors)", p.errors)
		}
		fmt.Fprintln(os.Stderr)
	}
}

// Summary prints the final compilation summary.
func (p *Progress) Summary(result *CompileResult) {
	elapsed := time.Since(p.startTime)

	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "📊 Compile complete in %s\n", elapsed.Round(time.Second))
	fmt.Fprintf(os.Stderr, "   Sources:  +%d added, ~%d modified, -%d removed\n",
		result.Added, result.Modified, result.Removed)
	if result.Summarized > 0 {
		fmt.Fprintf(os.Stderr, "   Summaries: %d written\n", result.Summarized)
	}
	if result.ConceptsExtracted > 0 {
		fmt.Fprintf(os.Stderr, "   Concepts:  %d extracted\n", result.ConceptsExtracted)
	}
	if result.ArticlesWritten > 0 {
		fmt.Fprintf(os.Stderr, "   Articles:  %d written\n", result.ArticlesWritten)
	}
	if result.Errors > 0 {
		fmt.Fprintf(os.Stderr, "   Errors:    %d\n", result.Errors)
	}
	fmt.Fprintln(os.Stderr)
}

func (p *Progress) progressBar() string {
	if p.total == 0 {
		return ""
	}
	width := 20
	filled := (p.done * width) / p.total
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return fmt.Sprintf("[%s] %d/%d", bar, p.done, p.total)
}

func truncatePath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	// Show the last maxLen-3 chars with "..." prefix
	return "..." + path[len(path)-maxLen+3:]
}
