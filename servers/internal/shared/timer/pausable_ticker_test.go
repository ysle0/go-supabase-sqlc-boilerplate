package timer

import (
	"fmt"
	"testing"
	"time"
)

func TestPausableCountdownTicker(t *testing.T) {
	// Test basic functionality
	duration := 5 * time.Second
	ticker := NewPausableCountdownTicker(duration)
	defer ticker.Stop()

	// Test getInitialTimeleft
	if ticker.GetInitialTimeleft() != duration {
		t.Errorf("Expected initial duration %v, got %v", duration, ticker.GetInitialTimeleft())
	}

	// Test getCurrentTimeleft (should be close to initial)
	remaining := ticker.GetCurrentTimeLeft()
	if remaining <= 0 || remaining > duration {
		t.Errorf("Expected remaining time between 0 and %v, got %v", duration, remaining)
	}

	// Wait a bit and check time decreased
	time.Sleep(500 * time.Millisecond)
	newRemaining := ticker.GetCurrentTimeLeft()
	if newRemaining >= remaining {
		t.Errorf("Expected time to decrease, was %v, now %v", remaining, newRemaining)
	}

	// Test pause functionality
	ticker.Pause()
	if !ticker.IsPaused() {
		t.Error("Expected ticker to be paused")
	}

	pausedRemaining := ticker.GetCurrentTimeLeft()
	time.Sleep(500 * time.Millisecond)
	stillPausedRemaining := ticker.GetCurrentTimeLeft()

	if pausedRemaining != stillPausedRemaining {
		t.Errorf("Expected time to remain same while paused, was %v, now %v", pausedRemaining, stillPausedRemaining)
	}

	// Test resume functionality
	ticker.Resume()
	if ticker.IsPaused() {
		t.Error("Expected ticker to be resumed")
	}

	time.Sleep(200 * time.Millisecond)
	resumedRemaining := ticker.GetCurrentTimeLeft()
	if resumedRemaining >= stillPausedRemaining {
		t.Errorf("Expected time to continue decreasing after resume, was %v, now %v", stillPausedRemaining, resumedRemaining)
	}
}

func TestPausableCountdownTickerChannelBehavior(t *testing.T) {
	ticker := NewPausableCountdownTicker(1 * time.Second)
	defer ticker.Stop()

	// Test that channel receives ticks
	select {
	case <-ticker.C:
		// Good, received a tick
	case <-time.After(200 * time.Millisecond):
		t.Error("Expected to receive tick within 200ms")
	}
}

func TestPausableCountdownTickerExpiration(t *testing.T) {
	ticker := NewPausableCountdownTicker(300 * time.Millisecond)
	defer ticker.Stop()

	// Wait for expiration
	time.Sleep(400 * time.Millisecond)

	if !ticker.IsExpired() {
		t.Error("Expected ticker to be expired")
	}

	if ticker.GetCurrentTimeLeft() != 0 {
		t.Errorf("Expected remaining time to be 0, got %v", ticker.GetCurrentTimeLeft())
	}
}

func TestPausableCountdownTickerReset(t *testing.T) {
	duration := 2 * time.Second
	ticker := NewPausableCountdownTicker(duration)
	defer ticker.Stop()

	// Wait a bit
	time.Sleep(300 * time.Millisecond)

	// Reset the timer
	ticker.Reset()

	// Check that it's back to initial duration
	remaining := ticker.GetCurrentTimeLeft()
	if remaining <= 0 || remaining > duration {
		t.Errorf("Expected remaining time after reset to be close to %v, got %v", duration, remaining)
	}

	if ticker.IsPaused() {
		t.Error("Expected ticker to be running after reset")
	}
}

// Example usage demonstration
func ExamplePausableCountdownTicker() {
	// Create a 3-second countdown timer
	countdown := NewPausableCountdownTicker(3 * time.Second)
	defer countdown.Stop()

	// Monitor the countdown
	go func() {
		for range countdown.C {
			remaining := countdown.GetCurrentTimeLeft()
			fmt.Printf("Time remaining: %v\n", remaining.Round(time.Millisecond))
		}
		fmt.Println("Countdown finished!")
	}()

	// Demonstrate pause and resume
	time.Sleep(1 * time.Second)
	fmt.Printf("Pausing at: %v\n", countdown.GetCurrentTimeLeft().Round(time.Millisecond))
	countdown.Pause()

	time.Sleep(1 * time.Second) // Timer paused, so this won't affect countdown
	fmt.Printf("Resuming at: %v\n", countdown.GetCurrentTimeLeft().Round(time.Millisecond))
	countdown.Resume()

	// Wait for completion
	time.Sleep(3 * time.Second)

	fmt.Printf("Initial duration was: %v\n", countdown.GetInitialTimeleft())
	fmt.Printf("Final remaining time: %v\n", countdown.GetCurrentTimeLeft())
}
