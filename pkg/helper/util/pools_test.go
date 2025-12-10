package util

import (
	"bytes"
	"sync"
	"testing"
)

func TestNewBufferPool(t *testing.T) {
	pool := NewBufferPool()
	if pool == nil {
		t.Error("Expected non-nil pool")
	}
}

func TestBufferPoolGetPut(t *testing.T) {
	pool := NewBufferPool()

	sizes := []int{512, 1024, 4096, 16384, 65536}
	for _, size := range sizes {
		buf, actualSize := pool.Get(size)
		if buf == nil {
			t.Errorf("Expected buffer for size %d", size)
		}
		if len(buf) != size {
			t.Errorf("Expected length %d, got %d", size, len(buf))
		}
		if actualSize < size {
			t.Errorf("Expected actual size >= %d, got %d", size, actualSize)
		}
		pool.Put(buf, actualSize)
	}
}

func TestBufferPoolReuse(t *testing.T) {
	pool := NewBufferPool()

	buf1, size := pool.Get(1024)
	for i := range buf1 {
		buf1[i] = byte(i % 256)
	}
	pool.Put(buf1, size)

	buf2, _ := pool.Get(1024)
	// Should be zeroed
	for i, b := range buf2[:100] {
		if b != 0 {
			t.Errorf("Expected zero at %d, got %d", i, b)
			break
		}
	}
}

func TestBufferPoolConcurrent(t *testing.T) {
	pool := NewBufferPool()
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buf, size := pool.Get(4096)
			pool.Put(buf, size)
		}()
	}
	wg.Wait()
}

func TestGlobalBufferPool(t *testing.T) {
	if GlobalBufferPool == nil {
		t.Error("Expected non-nil global pool")
	}
	buf, size := GlobalBufferPool.Get(1024)
	GlobalBufferPool.Put(buf, size)
}

func TestBytesBufferPool(t *testing.T) {
	pool := GlobalBytesBufferPool

	buf := pool.Get()
	if buf == nil {
		t.Fatal("Expected non-nil buffer")
	}
	if buf.Len() != 0 {
		t.Error("Expected empty buffer")
	}

	buf.WriteString("test")
	pool.Put(buf)

	buf2 := pool.Get()
	if buf2.Len() != 0 {
		t.Error("Expected reset buffer")
	}
}

func TestNewObjectPool(t *testing.T) {
	pool := NewObjectPool(func() interface{} {
		return &bytes.Buffer{}
	})

	obj := pool.Get()
	if obj == nil {
		t.Error("Expected non-nil object")
	}

	buf, ok := obj.(*bytes.Buffer)
	if !ok {
		t.Error("Expected *bytes.Buffer")
	}

	buf.WriteString("test")
	pool.Put(buf)
}

func TestNewMemoryOptimizer(t *testing.T) {
	opt := NewMemoryOptimizer()
	if opt == nil {
		t.Error("Expected non-nil optimizer")
	}

	src := []byte("source")
	dst := make([]byte, len(src))
	n := opt.OptimizedCopy(dst, src)
	if n != len(src) || string(dst) != string(src) {
		t.Error("Copy failed")
	}
}

func TestNewReusableBuffer(t *testing.T) {
	buf := NewReusableBuffer(1024)
	if buf == nil {
		t.Fatal("Expected non-nil buffer")
	}
	if buf.Len() != 1024 || buf.Cap() < 1024 {
		t.Error("Expected correct size")
	}

	// Test resize
	if err := buf.Resize(512); err != nil {
		t.Error("Resize failed")
	}
	if buf.Len() != 512 {
		t.Error("Expected length 512")
	}

	// Test resize beyond capacity
	if err := buf.Resize(buf.Cap() + 1); err != ErrBufferTooSmall {
		t.Error("Expected ErrBufferTooSmall")
	}

	buf.Release()
	// Multiple releases should be safe
	buf.Release()
}

func TestNewBufferManager(t *testing.T) {
	mgr := NewBufferManager()
	if mgr == nil {
		t.Error("Expected non-nil manager")
	}

	operations := []string{"compress", "network", "copy", "default"}
	for _, op := range operations {
		buf := mgr.GetOptimalBuffer(1024, op)
		if buf == nil {
			t.Errorf("Expected buffer for operation %s", op)
		}
		buf.Release()
	}
}
