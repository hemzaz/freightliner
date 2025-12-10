package util

import (
	"runtime"
	"testing"
	"time"
)

func TestDefaultGCOptimizerConfig(t *testing.T) {
	cfg := DefaultGCOptimizerConfig()
	if cfg.LowMemoryGCPercent <= 0 || cfg.HighMemoryGCPercent <= 0 {
		t.Error("Expected positive GC percentages")
	}
}

func TestNewGCOptimizer(t *testing.T) {
	cfg := DefaultGCOptimizerConfig()
	opt := NewGCOptimizer(cfg, nil)
	if opt == nil {
		t.Fatal("Expected non-nil optimizer")
	}
}

func TestGCOptimizerStartStop(t *testing.T) {
	cfg := DefaultGCOptimizerConfig()
	cfg.MonitoringInterval = 50 * time.Millisecond
	cfg.GCStatsInterval = 100 * time.Millisecond
	opt := NewGCOptimizer(cfg, nil)

	opt.Start()
	time.Sleep(60 * time.Millisecond)
	opt.Stop()
}

func TestGCOptimizerForceGC(t *testing.T) {
	cfg := DefaultGCOptimizerConfig()
	opt := NewGCOptimizer(cfg, nil)

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	// Allocate
	_ = make([]byte, 1024*1024)

	opt.ForceGC("test")

	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	if after.NumGC <= before.NumGC {
		t.Error("Expected GC to run")
	}
}

func TestGCOptimizerGetStats(t *testing.T) {
	cfg := DefaultGCOptimizerConfig()
	opt := NewGCOptimizer(cfg, nil)

	stats := opt.GetStats()
	if stats == nil || stats.GCStats == nil {
		t.Error("Expected non-nil stats")
	}
}

func TestObjectLifecycleWrapper(t *testing.T) {
	wrapper := &ObjectLifecycleWrapper{}

	if wrapper.IsInUse() {
		t.Error("Expected not in use initially")
	}

	wrapper.Get()
	if !wrapper.IsInUse() {
		t.Error("Expected in use after Get()")
	}

	wrapper.Release()
	if wrapper.IsInUse() {
		t.Error("Expected not in use after Release()")
	}

	age := wrapper.Age()
	if age < 0 {
		t.Error("Expected non-negative age")
	}
}

func TestGCOptimizerGetReturnObject(t *testing.T) {
	cfg := DefaultGCOptimizerConfig()
	opt := NewGCOptimizer(cfg, nil)

	for _, cat := range []ObjectLifecycleCategory{ShortLived, MediumLived, LongLived} {
		wrapper := opt.GetObject(cat)
		if wrapper == nil {
			t.Errorf("Expected wrapper for category %d", cat)
		}
		opt.ReturnObject(wrapper)
	}
}

func TestGlobalGCOptimizer(t *testing.T) {
	if GlobalGCOptimizer == nil {
		t.Error("Expected non-nil global optimizer")
	}
}

func TestOptimizeForContainerRegistry(t *testing.T) {
	opt := OptimizeForContainerRegistry()
	if opt == nil {
		t.Error("Expected non-nil optimizer")
	}
}

func TestNewMemoryEfficientProcessor(t *testing.T) {
	cfg := DefaultGCOptimizerConfig()
	opt := NewGCOptimizer(cfg, nil)
	proc := NewMemoryEfficientProcessor(opt, nil)
	if proc == nil {
		t.Error("Expected non-nil processor")
	}
}

func TestMemoryEfficientProcessorProcessWithMinimalAllocation(t *testing.T) {
	cfg := DefaultGCOptimizerConfig()
	opt := NewGCOptimizer(cfg, nil)
	proc := NewMemoryEfficientProcessor(opt, nil)

	data := []byte("test")
	result, err := proc.ProcessWithMinimalAllocation(data, func(d []byte) ([]byte, error) {
		return d, nil
	})
	if err != nil || len(result) == 0 {
		t.Error("Expected successful processing")
	}
}

func TestMemoryEfficientProcessorBatchProcess(t *testing.T) {
	cfg := DefaultGCOptimizerConfig()
	opt := NewGCOptimizer(cfg, nil)
	opt.Start()
	defer opt.Stop()

	proc := NewMemoryEfficientProcessor(opt, nil)

	items := make([]interface{}, 20)
	for i := range items {
		items[i] = i
	}

	count := 0
	err := proc.BatchProcessWithGCControl(items, func(item interface{}) error {
		count++
		return nil
	}, 5)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if count != len(items) {
		t.Errorf("Expected to process %d items, got %d", len(items), count)
	}
}

func TestObjectSizeCalculator(t *testing.T) {
	calc := &ObjectSizeCalculator{}

	size1 := calc.SizeOf("test")
	if size1 == 0 {
		t.Error("Expected non-zero size for string")
	}

	size2 := calc.SizeOf([]byte{1, 2, 3})
	if size2 == 0 {
		t.Error("Expected non-zero size for byte slice")
	}

	size3 := calc.SizeOf([]interface{}{1, "test"})
	if size3 == 0 {
		t.Error("Expected non-zero size for interface slice")
	}

	size4 := calc.SizeOf(42)
	if size4 == 0 {
		t.Error("Expected non-zero size for int")
	}
}
