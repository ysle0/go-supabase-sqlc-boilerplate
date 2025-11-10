package timer

import (
	"time"
)

type PausableCountdownTicker struct {
	C               <-chan time.Time
	initialDuration time.Duration
	remaining       time.Duration
	lastStartTime   time.Time
	paused          bool
	ticker          *time.Ticker
	done            chan struct{}
	output          chan time.Time
}

func NewPausableCountdownTicker(d time.Duration) *PausableCountdownTicker {
	if d <= 0 {
		d = time.Millisecond // Prevent zero or negative durations
	}

	output := make(chan time.Time, 1)
	tickerInterval := time.Millisecond * 100 // Fixed 100ms intervals for smooth countdown

	pct := &PausableCountdownTicker{
		C:               output,
		initialDuration: d,
		remaining:       d,
		lastStartTime:   time.Now(),
		paused:          false,
		ticker:          time.NewTicker(tickerInterval),
		done:            make(chan struct{}),
		output:          output,
	}

	go pct.run()
	return pct
}

func (pct *PausableCountdownTicker) run() {
	defer close(pct.output)
	defer pct.ticker.Stop()

	for {
		select {
		case t := <-pct.ticker.C:
			if !pct.paused {
				elapsed := time.Since(pct.lastStartTime)
				pct.remaining -= elapsed
				pct.lastStartTime = time.Now()

				if pct.remaining <= 0 {
					pct.remaining = 0
					// Send final tick and exit
					pct.output <- t
					pct.done <- struct{}{}

					return
				}

				// Send tick
				select {
				case pct.output <- t:
				default:
				}
			}

		case <-pct.done:
			return
		}
	}
}

// Pause pauses the countdown timer
func (pct *PausableCountdownTicker) Pause() {
	if !pct.paused {
		// Update the remaining time before pausing
		elapsed := time.Since(pct.lastStartTime)
		pct.remaining -= elapsed
		if pct.remaining < 0 {
			pct.remaining = 0
		}
		pct.paused = true
	}
}

// Resume resumes the countdown timer
func (pct *PausableCountdownTicker) Resume() {
	if pct.paused {
		pct.lastStartTime = time.Now()
		pct.paused = false
	}
}

// GetCurrentTimeLeft returns the current remaining time
func (pct *PausableCountdownTicker) GetCurrentTimeLeft() time.Duration {
	if pct.paused {
		return pct.remaining
	}

	elapsed := time.Since(pct.lastStartTime)
	current := pct.remaining - elapsed
	if current < 0 {
		return 0
	}
	return current
}

// GetInitialTimeleft returns the original duration set at creation
func (pct *PausableCountdownTicker) GetInitialTimeleft() time.Duration {
	return pct.initialDuration
}

// Stop stops the countdown timer completely
func (pct *PausableCountdownTicker) Stop() {
	select {
	case <-pct.done:
		// Already stopped
		return
	default:
		close(pct.done)
		pct.remaining = 0
		pct.paused = true
	}
}

// IsExpired returns true if the countdown has reached zero
func (pct *PausableCountdownTicker) IsExpired() bool {
	return pct.GetCurrentTimeLeft() == 0
}

// IsPaused returns true if the countdown is currently paused
func (pct *PausableCountdownTicker) IsPaused() bool {
	return pct.paused
}

// Reset resets the countdown timer to its initial duration
func (pct *PausableCountdownTicker) Reset() {
	pct.remaining = pct.initialDuration
	pct.lastStartTime = time.Now()
	pct.paused = false
}

// Legacy compatibility - keeping the old interface for backward compatibility
type PausableTicker = PausableCountdownTicker

func NewPausableTicker(d time.Duration) *PausableTicker {
	return NewPausableCountdownTicker(d)
}

// Legacy methods for backward compatibility
func (pct *PausableCountdownTicker) Timeleft() time.Duration {
	return pct.GetCurrentTimeLeft()
}
