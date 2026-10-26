package ui

import (
	"regexp"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const (
	maxLogLines        = 1000
	uiThrottleInterval = 80 * time.Millisecond
)

var ansiEscapePattern = regexp.MustCompile(`\x1b\[[0-?]*[ -/]*[@-~]`)

func stripANSI(text string) string {
	return ansiEscapePattern.ReplaceAllString(text, "")
}

// LogView manages the scrollable terminal output grid with line limits and UI throttling.
type LogView struct {
	Grid   *widget.TextGrid
	Scroll *container.Scroll

	mu              sync.Mutex
	pendingBuf      strings.Builder
	logLines        []string
	incompleteLine  string
	logText         string
	uiUpdatePending bool
	updateTimer     *time.Timer
}

func newLogView() *LogView {
	grid := widget.NewTextGrid()
	grid.SetText("Waiting.\n")
	scroll := container.NewScroll(grid)
	scroll.SetMinSize(fyne.NewSize(760, 260))

	return &LogView{
		Grid:   grid,
		Scroll: scroll,
	}
}

func (lv *LogView) applyPendingLocked(flushIncomplete bool) string {
	if lv.pendingBuf.Len() > 0 || (flushIncomplete && lv.incompleteLine != "") {
		chunk := lv.pendingBuf.String()
		lv.pendingBuf.Reset()

		combined := lv.incompleteLine + chunk
		parts := strings.Split(combined, "\n")
		if flushIncomplete {
			lv.incompleteLine = ""
			for i, p := range parts {
				if i == len(parts)-1 && p == "" {
					continue
				}
				lv.logLines = append(lv.logLines, p)
			}
		} else {
			lv.incompleteLine = parts[len(parts)-1]
			if len(parts) > 1 {
				lv.logLines = append(lv.logLines, parts[:len(parts)-1]...)
			}
		}

		if len(lv.logLines) > maxLogLines {
			trimmed := make([]string, maxLogLines)
			copy(trimmed, lv.logLines[len(lv.logLines)-maxLogLines:])
			lv.logLines = trimmed
		}

		if lv.incompleteLine != "" {
			if len(lv.logLines) > 0 {
				lv.logText = strings.Join(lv.logLines, "\n") + "\n" + lv.incompleteLine
			} else {
				lv.logText = lv.incompleteLine
			}
		} else {
			lv.logText = strings.Join(lv.logLines, "\n")
		}
	}
	return lv.logText
}

func (lv *LogView) render() {
	lv.mu.Lock()
	if lv.updateTimer != nil {
		lv.updateTimer.Stop()
		lv.updateTimer = nil
	}
	lv.uiUpdatePending = false
	current := lv.applyPendingLocked(false)
	lv.mu.Unlock()

	lv.Grid.SetText(current)
	lv.Scroll.ScrollToBottom()
}

// Append strips ANSI escape codes and queues text to be displayed in the log grid.
func (lv *LogView) Append(text string) {
	clean := stripANSI(text)
	if clean == "" {
		return
	}
	lv.mu.Lock()
	lv.pendingBuf.WriteString(clean)
	if !lv.uiUpdatePending {
		lv.uiUpdatePending = true
		lv.updateTimer = time.AfterFunc(uiThrottleInterval, func() {
			fyne.Do(lv.render)
		})
	}
	lv.mu.Unlock()
}

// Flush forces any pending output to be rendered on the UI thread before returning.
func (lv *LogView) Flush() {
	done := make(chan struct{})
	fyne.Do(func() {
		lv.mu.Lock()
		if lv.updateTimer != nil {
			lv.updateTimer.Stop()
			lv.updateTimer = nil
		}
		lv.uiUpdatePending = false
		current := lv.applyPendingLocked(true)
		lv.mu.Unlock()

		lv.Grid.SetText(current)
		lv.Scroll.ScrollToBottom()
		close(done)
	})
	<-done
}

// Reset clears the buffered and displayed log.
func (lv *LogView) Reset() {
	lv.mu.Lock()
	if lv.updateTimer != nil {
		lv.updateTimer.Stop()
		lv.updateTimer = nil
	}
	lv.uiUpdatePending = false
	lv.pendingBuf.Reset()
	lv.logLines = nil
	lv.incompleteLine = ""
	lv.logText = ""
	lv.mu.Unlock()
	fyne.Do(func() {
		lv.Grid.SetText("")
	})
}

// Text returns the full current log text.
func (lv *LogView) Text() string {
	lv.mu.Lock()
	defer lv.mu.Unlock()
	return lv.applyPendingLocked(false)
}
