package util

import (
	"sync"
	"testing"
	"time"
)

// Performance Monitor Tests
func TestNewPerformanceMonitor(t *testing.T) {
	mon := NewPerformanceMonitor(nil)
	if mon == nil {
		t.Fatal("Expected non-nil monitor")
	}
}

func TestPerformanceMonitorStartStop(t *testing.T) {
	mon := NewPerformanceMonitor(nil)
	mon.Start()
	time.Sleep(100 * time.Millisecond)
	mon.Stop()
	// Multiple stops should be safe
	mon.Stop()
}

func TestOperationTracker(t *testing.T) {
	mon := NewPerformanceMonitor(nil)
	tracker := mon.StartOperation("test_op")
	if tracker == nil {
		t.Fatal("Expected non-nil tracker")
	}

	tracker.AddBytes(1024)
	tracker.AddItems(10)
	time.Sleep(5 * time.Millisecond)
	tracker.Finish(nil)

	metrics, exists := mon.GetOperationMetrics("test_op")
	if !exists || metrics.Count.Load() != 1 {
		t.Error("Expected operation to be recorded")
	}
	if metrics.BytesProcessed.Load() != 1024 {
		t.Error("Expected bytes to be recorded")
	}
	if metrics.ItemsProcessed.Load() != 10 {
		t.Error("Expected items to be recorded")
	}
}

func TestPerformanceMonitorGetAllOperationMetrics(t *testing.T) {
	mon := NewPerformanceMonitor(nil)
	mon.StartOperation("op1").Finish(nil)
	mon.StartOperation("op2").Finish(nil)

	all := mon.GetAllOperationMetrics()
	if len(all) != 2 {
		t.Errorf("Expected 2 operations, got %d", len(all))
	}
}

func TestPerformanceMonitorGenerateReport(t *testing.T) {
	mon := NewPerformanceMonitor(nil)
	mon.Start()
	defer mon.Stop()

	tracker := mon.StartOperation("test")
	time.Sleep(5 * time.Millisecond)
	tracker.Finish(nil)

	time.Sleep(50 * time.Millisecond)

	report := mon.GenerateReport()
	if report == nil {
		t.Fatal("Expected non-nil report")
	}
	if report.SystemMetrics.CPUCores <= 0 {
		t.Error("Expected positive CPU cores")
	}
}

func TestGlobalPerformanceMonitor(t *testing.T) {
	if GlobalPerformanceMonitor == nil {
		t.Error("Expected non-nil global monitor")
	}
}

func TestNewLatencyHistogram(t *testing.T) {
	h := NewLatencyHistogram()
	if h == nil {
		t.Fatal("Expected non-nil histogram")
	}

	h.Record(500 * time.Microsecond)
	h.Record(5 * time.Millisecond)
	h.Record(50 * time.Millisecond)
	h.Record(500 * time.Millisecond)
	h.Record(5 * time.Second)
	h.Record(15 * time.Second)

	dist := h.GetDistribution()
	if len(dist) != 6 {
		t.Errorf("Expected 6 buckets, got %d", len(dist))
	}
}

func TestNewBenchmarkSuite(t *testing.T) {
	suite := NewBenchmarkSuite(nil)
	if suite == nil {
		t.Fatal("Expected non-nil suite")
	}

	executed := 0
	result := suite.RunBenchmark("test", 10, func() error {
		executed++
		time.Sleep(1 * time.Millisecond)
		return nil
	})

	if executed != 10 {
		t.Errorf("Expected 10 executions, got %d", executed)
	}
	if result == nil || result.Iterations != 10 {
		t.Error("Expected valid result")
	}
}

func TestBenchmarkSuiteGetResults(t *testing.T) {
	suite := NewBenchmarkSuite(nil)
	suite.RunBenchmark("b1", 5, func() error { return nil })
	suite.RunBenchmark("b2", 5, func() error { return nil })

	results := suite.GetBenchmarkResults()
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

// CPU Algorithm Tests
func TestNewCPUEfficientSorter(t *testing.T) {
	sorter := NewCPUEfficientSorter(nil)
	if sorter == nil {
		t.Error("Expected non-nil sorter")
	}
}

func TestSortResourcesByPriority(t *testing.T) {
	sorter := NewCPUEfficientSorter(nil)

	resources := []SortableResource{
		{Name: "low", Priority: 1},
		{Name: "high", Priority: 10},
		{Name: "med", Priority: 5},
	}

	sorter.SortResourcesByPriority(resources)

	if resources[0].Priority != 10 || resources[1].Priority != 5 || resources[2].Priority != 1 {
		t.Error("Expected sorted by priority (high to low)")
	}
}

func TestNewStringMatcher(t *testing.T) {
	m := NewStringMatcher("pattern")
	if m == nil {
		t.Error("Expected non-nil matcher")
	}

	matches := m.FindAll("this pattern is a pattern test")
	if len(matches) != 2 {
		t.Errorf("Expected 2 matches, got %d", len(matches))
	}

	complexity := m.GetComplexity()
	if complexity == "" {
		t.Error("Expected non-empty complexity")
	}
}

func TestNewEfficientSetOperations(t *testing.T) {
	ops := NewEfficientSetOperations[int](nil)
	if ops == nil {
		t.Error("Expected non-nil ops")
	}

	// Test intersection
	set1 := []int{1, 2, 3, 4}
	set2 := []int{3, 4, 5, 6}
	intersection := ops.Intersection(set1, set2)
	if len(intersection) != 2 {
		t.Errorf("Expected 2 elements in intersection, got %d", len(intersection))
	}

	// Test union
	union := ops.Union(set1, set2)
	if len(union) != 6 {
		t.Errorf("Expected 6 elements in union, got %d", len(union))
	}

	// Test difference
	diff := ops.Difference(set1, set2)
	if len(diff) != 2 {
		t.Errorf("Expected 2 elements in difference, got %d", len(diff))
	}
}

func TestNewPatternMatchingCache(t *testing.T) {
	cache := NewPatternMatchingCache(10, nil)
	if cache == nil {
		t.Error("Expected non-nil cache")
	}

	if !cache.Match("test", "this is a test") {
		t.Error("Expected match")
	}
	if cache.Match("xyz", "abc") {
		t.Error("Expected no match")
	}
}

func TestNewBinarySearchTree(t *testing.T) {
	compare := func(a, b int) int {
		if a < b {
			return -1
		}
		if a > b {
			return 1
		}
		return 0
	}

	tree := NewBinarySearchTree[int](compare)
	if tree == nil {
		t.Error("Expected non-nil tree")
	}

	values := []int{5, 3, 7, 1, 9}
	for _, v := range values {
		tree.Insert(v)
	}

	if tree.Size() != len(values) {
		t.Errorf("Expected size %d, got %d", len(values), tree.Size())
	}

	for _, v := range values {
		if !tree.Search(v) {
			t.Errorf("Expected to find %d", v)
		}
	}

	if tree.Search(100) {
		t.Error("Expected not to find 100")
	}
}

func TestNewHashTable(t *testing.T) {
	table := NewHashTable[string, int](16)
	if table == nil {
		t.Error("Expected non-nil table")
	}

	table.Put("key1", 100)
	value, exists := table.Get("key1")
	if !exists || value != 100 {
		t.Error("Expected to find key1 with value 100")
	}

	table.Put("key1", 200)
	value, _ = table.Get("key1")
	if value != 200 {
		t.Error("Expected updated value 200")
	}

	_, exists = table.Get("nonexistent")
	if exists {
		t.Error("Expected key not to exist")
	}
}

func TestNewAlgorithmProfiler(t *testing.T) {
	profiler := NewAlgorithmProfiler(nil)
	if profiler == nil {
		t.Error("Expected non-nil profiler")
	}

	executed := false
	profiler.ProfileAlgorithm("test", ComplexityON, 100, func() {
		executed = true
	})

	if !executed {
		t.Error("Expected algorithm to execute")
	}

	profile, exists := profiler.GetProfile("test")
	if !exists || profile.ExecutionCount != 1 {
		t.Error("Expected profile to be recorded")
	}

	all := profiler.GetAllProfiles()
	if len(all) != 1 {
		t.Error("Expected 1 profile")
	}
}

func TestGlobalInstances(t *testing.T) {
	if GlobalCPUSorter == nil {
		t.Error("Expected non-nil GlobalCPUSorter")
	}
	if GlobalPatternCache == nil {
		t.Error("Expected non-nil GlobalPatternCache")
	}
	if GlobalAlgProfiler == nil {
		t.Error("Expected non-nil GlobalAlgProfiler")
	}
}

func TestConcurrentOperationTracking(t *testing.T) {
	mon := NewPerformanceMonitor(nil)
	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tracker := mon.StartOperation("concurrent")
			tracker.AddBytes(100)
			time.Sleep(1 * time.Millisecond)
			tracker.Finish(nil)
		}()
	}

	wg.Wait()

	metrics, exists := mon.GetOperationMetrics("concurrent")
	if !exists || metrics.Count.Load() != 20 {
		t.Error("Expected 20 operations")
	}
}
