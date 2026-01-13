package main

import (
	"testing"
	"time"
)

// Unit tests for graceful shutdown (Story 7.7)

// TestShutdownTimeout_Value verifies the shutdown timeout matches architecture spec.
// AC2: 5-second timeout per architecture.md specification.
func TestShutdownTimeout_Value(t *testing.T) {
	expected := 5 * time.Second
	if shutdownTimeout != expected {
		t.Errorf("shutdownTimeout = %v, want %v", shutdownTimeout, expected)
	}
}

// TestShutdownTimeout_NotZero verifies timeout is not zero.
func TestShutdownTimeout_NotZero(t *testing.T) {
	if shutdownTimeout == 0 {
		t.Error("shutdownTimeout should not be zero")
	}
}

// TestShutdownTimeout_Reasonable verifies timeout is within reasonable bounds.
// Too short: cleanup might not complete
// Too long: user will think app is stuck
func TestShutdownTimeout_Reasonable(t *testing.T) {
	if shutdownTimeout < 1*time.Second {
		t.Error("shutdownTimeout is too short (< 1s)")
	}
	if shutdownTimeout > 30*time.Second {
		t.Error("shutdownTimeout is too long (> 30s)")
	}
}

// TestDoneChannel_Behavior verifies the done channel pattern works correctly.
// This tests the core synchronization mechanism for graceful shutdown.
func TestDoneChannel_Behavior(t *testing.T) {
	done := make(chan struct{})

	// Simulate signal handler waiting on done
	signalHandlerDone := make(chan struct{})
	go func() {
		select {
		case <-done:
			// Expected path - done channel closed
		case <-time.After(100 * time.Millisecond):
			t.Error("Signal handler timed out waiting for done channel")
		}
		close(signalHandlerDone)
	}()

	// Simulate run() completing and closing done
	close(done)

	// Verify signal handler exits promptly
	select {
	case <-signalHandlerDone:
		// Success - signal handler noticed done channel
	case <-time.After(200 * time.Millisecond):
		t.Error("Signal handler didn't exit after done channel closed")
	}
}

// TestDoneChannel_MultipleReaders verifies done channel works with multiple readers.
// This is relevant because both main and the signal goroutine interact with done.
func TestDoneChannel_MultipleReaders(t *testing.T) {
	done := make(chan struct{})

	// Multiple goroutines waiting
	const numReaders = 3
	allDone := make(chan struct{})
	doneCount := make(chan struct{}, numReaders)

	for i := 0; i < numReaders; i++ {
		go func() {
			<-done
			doneCount <- struct{}{}
		}()
	}

	// Close done and verify all readers wake up
	close(done)

	go func() {
		for i := 0; i < numReaders; i++ {
			<-doneCount
		}
		close(allDone)
	}()

	select {
	case <-allDone:
		// All readers got the signal
	case <-time.After(100 * time.Millisecond):
		t.Error("Not all readers received done signal")
	}
}
