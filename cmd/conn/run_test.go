package main

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockActivityService struct {
	syncCalls int
	syncFn    func(ctx context.Context) error
}

func (m *mockActivityService) Sync(ctx context.Context) error {
	m.syncCalls++
	if m.syncFn != nil {
		return m.syncFn(ctx)
	}
	return nil
}

// Test: interval <= 0 returns 0
func TestRun_InvalidInterval(t *testing.T) {
	ctx := context.Background()
	mock := &mockActivityService{}

	err := run(ctx, mock, 0)
	if err == nil {
		t.Fatalf("expected error for interval <= 0, got nil")
	}
}

// Test: interval Sync is called once on startup
func TestRun_InitialSync(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mock := &mockActivityService{}
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	_ = run(ctx, mock, 50*time.Millisecond)

	if mock.syncCalls != 1 {
		t.Fatalf("expected initial Sync call = 1, got %d", mock.syncCalls)
	}
}

func TestRun_ErrOnInitialSync(t *testing.T) {
	ctx := context.Background()

	mock := &mockActivityService{
		syncFn: func(ctx context.Context) error {
			return errors.New("test sync error")
		},
	}

	err := run(ctx, mock, 10*time.Millisecond)
	if err == nil {
		t.Fatalf("expected error from initial sync, got nil")
	}

	if mock.syncCalls > 1 {
		t.Fatalf("Sync should only be called exactly once before error, got %d sync calls", mock.syncCalls)
	}
}

func TestRun_TickerSync(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mock := &mockActivityService{}

	go func() {
		time.Sleep(60 * time.Millisecond)
		cancel()
	}()

	_ = run(ctx, mock, 10*time.Millisecond)
	if mock.syncCalls < 2 {
		t.Fatalf("expected multiple Sync calls, got %d", mock.syncCalls)
	}
}

func TestRun_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	mock := &mockActivityService{}

	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	err := run(ctx, mock, 20*time.Millisecond)
	if err != nil {
		t.Fatalf("expected error on clean exit, go %v", err)
	}
}

func TestRun_SyncErrorDoesNotStop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	callCount := 0

	mock := &mockActivityService{
		syncFn: func(ctx context.Context) error {
			callCount++

			if callCount == 1 {
				return nil
			}
			return errors.New("test sync error")
		},
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := run(ctx, mock, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("run() should not return error from Sync, got %v", err)
	}

	if callCount < 2 {
		t.Fatalf("expected run() to continue after calling Sync after initial sync success; got %d calls", callCount)
	}
}
