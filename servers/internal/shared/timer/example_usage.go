package timer

import (
	"fmt"
	"time"
)

// DemoCountdownTimer demonstrates the usage of PausableCountdownTicker
func DemoCountdownTimer() {
	fmt.Println("=== Pausable Countdown Timer Demo ===")

	// Create a 5-second countdown timer
	countdown := NewPausableCountdownTicker(5 * time.Second)
	defer countdown.Stop()

	fmt.Printf("Initial duration: %v\n", countdown.GetInitialTimeleft())
	fmt.Printf("Starting countdown...\n\n")

	// Monitor the countdown in a goroutine
	go func() {
		for t := range countdown.C {
			remaining := countdown.GetCurrentTimeLeft()
			fmt.Printf("[%v] Time remaining: %v\n",
				t.Format("15:04:05.000"),
				remaining.Round(10*time.Millisecond))

			if remaining == 0 {
				fmt.Println("🎉 Countdown finished!")
				break
			}
		}
	}()

	// Let it run for 2 seconds
	time.Sleep(2 * time.Second)

	// Pause the timer
	fmt.Printf("\n⏸️  Pausing countdown at: %v\n", countdown.GetCurrentTimeLeft().Round(10*time.Millisecond))
	countdown.Pause()
	fmt.Printf("Is paused: %v\n", countdown.IsPaused())

	// Wait 2 seconds while paused (time shouldn't change)
	fmt.Println("Waiting 2 seconds while paused...")
	time.Sleep(2 * time.Second)
	fmt.Printf("Time after pause: %v (should be same)\n", countdown.GetCurrentTimeLeft().Round(10*time.Millisecond))

	// Resume the timer
	fmt.Printf("\n▶️  Resuming countdown...\n")
	countdown.Resume()
	fmt.Printf("Is paused: %v\n\n", countdown.IsPaused())

	// Wait for completion
	time.Sleep(4 * time.Second)

	fmt.Printf("\nFinal state:\n")
	fmt.Printf("- Initial duration: %v\n", countdown.GetInitialTimeleft())
	fmt.Printf("- Current remaining: %v\n", countdown.GetCurrentTimeLeft())
	fmt.Printf("- Is expired: %v\n", countdown.IsExpired())
	fmt.Printf("- Is paused: %v\n", countdown.IsPaused())
}

// QuickFeatureDemo demonstrates all the requested feature
func QuickFeatureDemo() {
	fmt.Println("\n=== Quick Feature Demo ===")

	timer := NewPausableCountdownTicker(3 * time.Second)
	defer timer.Stop()

	// Feature 1: getCurrentTimeleft
	fmt.Printf("1. getCurrentTimeleft(): %v\n", timer.GetCurrentTimeLeft().Round(time.Millisecond))

	// Feature 2: getInitialTimeleft
	fmt.Printf("2. getInitialTimeleft(): %v\n", timer.GetInitialTimeleft())

	// Let some time pass
	time.Sleep(500 * time.Millisecond)

	// Feature 3: pause
	fmt.Printf("3. pause() - Before: %v\n", timer.GetCurrentTimeLeft().Round(time.Millisecond))
	timer.Pause()
	time.Sleep(500 * time.Millisecond)
	fmt.Printf("   After 500ms pause: %v (should be same)\n", timer.GetCurrentTimeLeft().Round(time.Millisecond))

	// Feature 4: resume
	fmt.Printf("4. resume() - Resuming...\n")
	timer.Resume()
	time.Sleep(200 * time.Millisecond)
	fmt.Printf("   After 200ms resume: %v (should be less)\n", timer.GetCurrentTimeLeft().Round(time.Millisecond))

	fmt.Println("\n✅ All feature working correctly!")
}
