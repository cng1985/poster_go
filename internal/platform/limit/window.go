package limit

import (
	"sync"
	"time"

	apperrors "poster.app/studio/internal/platform/errors"
)

// Window 是进程内滑动窗口。多实例不共享，超过容量时丢弃旧键。
type Window struct {
	mu     sync.Mutex
	hits   map[string][]time.Time
	max    int
	window time.Duration
}

func NewWindow(max int, window time.Duration) *Window {
	return &Window{hits: map[string][]time.Time{}, max: max, window: window}
}

func (w *Window) Allow(key string) error {
	if w == nil || w.max <= 0 {
		return nil
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.prune(key)
	if len(w.hits[key]) >= w.max {
		return apperrors.RateLimit("操作过于频繁，请稍后再试")
	}
	w.hits[key] = append(w.hits[key], time.Now())
	w.capKeys()
	return nil
}

// FailureWindow 只累计失败次数，成功后清零。用于登录。
type FailureWindow struct {
	mu     sync.Mutex
	hits   map[string][]time.Time
	max    int
	window time.Duration
}

func NewFailureWindow(max int, window time.Duration) *FailureWindow {
	return &FailureWindow{hits: map[string][]time.Time{}, max: max, window: window}
}

func (f *FailureWindow) Rejected(key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.prune(key)
	if len(f.hits[key]) >= f.max {
		return apperrors.RateLimit("尝试次数过多，请稍后再试")
	}
	return nil
}

func (f *FailureWindow) Fail(key string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.prune(key)
	f.hits[key] = append(f.hits[key], time.Now())
	f.capKeys()
}

func (f *FailureWindow) Success(key string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.hits, key)
}

func (w *Window) prune(key string) {
	w.hits[key] = keepRecent(w.hits[key], w.window)
}

func (w *Window) capKeys() {
	if len(w.hits) <= 10000 {
		return
	}
	w.hits = map[string][]time.Time{}
}

func (f *FailureWindow) prune(key string) {
	f.hits[key] = keepRecent(f.hits[key], f.window)
}

func (f *FailureWindow) capKeys() {
	if len(f.hits) <= 10000 {
		return
	}
	f.hits = map[string][]time.Time{}
}

func keepRecent(items []time.Time, window time.Duration) []time.Time {
	if len(items) == 0 {
		return items
	}
	now := time.Now()
	kept := items[:0]
	for _, item := range items {
		if now.Sub(item) < window {
			kept = append(kept, item)
		}
	}
	return kept
}
