package timer

import (
	"log"
	"sync"
	"time"
)

type PausableTimer struct {
	mtx       sync.Mutex
	duration  time.Duration
	startedAt time.Time
	timer     *time.Timer
	deadline  time.Time
	remaining time.Duration
	paused    bool
}

func NewPausableTimer(d time.Duration) *PausableTimer {
	return &PausableTimer{
		duration:  d,
		startedAt: time.Now(),
		timer:     time.NewTimer(d),
		deadline:  time.Now().Add(d),
	}
}

func (pt *PausableTimer) Timeleft() time.Duration {
	elapsed := time.Since(pt.startedAt)
	remaining := pt.duration - elapsed
	log.Printf("remaining: %v", remaining)

	if remaining < 0 {
		remaining = 0
	}

	return remaining
}

func (pt *PausableTimer) C() <-chan time.Time {
	return pt.timer.C
}

func (pt *PausableTimer) Pause(latency time.Duration) time.Duration {
	pt.mtx.Lock()
	defer pt.mtx.Unlock()

	if pt.paused {
		return pt.remaining + latency
	}
	if !pt.timer.Stop() { // drain if already fired
		select {
		case <-pt.timer.C:
		default:
		}
	}
	pt.remaining = time.Until(pt.deadline) + latency
	pt.deadline = pt.deadline.Add(latency)
	pt.paused = true
	return pt.remaining + latency
}

func (pt *PausableTimer) Resume() {
	pt.mtx.Lock()
	defer pt.mtx.Unlock()

	if !pt.paused {
		return
	}

	pt.timer.Reset(pt.remaining)
	pt.deadline = time.Now().Add(pt.remaining)
	pt.paused = false
}

func (pt *PausableTimer) Stop() {
	pt.timer.Stop()
	pt.remaining = 0
	pt.paused = true
}
