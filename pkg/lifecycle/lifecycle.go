package lifecycle

import (
	"context"
	"fmt"
	"time"
)

type LifecycleHook interface {
	OnStartup() error
	OnShutdown() error
}

type Lifecycle struct {
	hooks           []LifecycleHook
	startupTimeout  time.Duration
	shutdownTimeout time.Duration
}

func NewLifecycle() *Lifecycle {
	return &Lifecycle{
		hooks:           make([]LifecycleHook, 0),
		startupTimeout:  30 * time.Second,
		shutdownTimeout: 10 * time.Second,
	}
}

func (l *Lifecycle) SetStartupTimeout(d time.Duration) *Lifecycle {
	l.startupTimeout = d
	return l
}

func (l *Lifecycle) SetShutdownTimeout(d time.Duration) *Lifecycle {
	l.shutdownTimeout = d
	return l
}

func (l *Lifecycle) Register(hook LifecycleHook) *Lifecycle {
	l.hooks = append(l.hooks, hook)
	return l
}

func (l *Lifecycle) ExecuteStartup() error {
	if len(l.hooks) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), l.startupTimeout)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		for _, hook := range l.hooks {
			if err := hook.OnStartup(); err != nil {
				errChan <- fmt.Errorf("startup hook failed: %w", err)
				return
			}
		}
		errChan <- nil
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return fmt.Errorf("startup timeout after %v", l.startupTimeout)
	}
}

func (l *Lifecycle) ExecuteShutdown() error {
	if len(l.hooks) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), l.shutdownTimeout)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		for i := len(l.hooks) - 1; i >= 0; i-- {
			if err := l.hooks[i].OnShutdown(); err != nil {
				errChan <- fmt.Errorf("shutdown hook failed: %w", err)
				return
			}
		}
		errChan <- nil
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return fmt.Errorf("shutdown timeout after %v", l.shutdownTimeout)
	}
}
