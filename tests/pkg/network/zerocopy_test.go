package network

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"freightliner/pkg/network"
)

func TestZeroCopy_CopyWithZeroCopy(t *testing.T) {
	data := []byte("test data for zero-copy transfer")
	src := bytes.NewReader(data)
	dst := &bytes.Buffer{}

	n, err := network.CopyWithZeroCopy(dst, src)
	if err != nil {
		t.Fatalf("copy failed: %v", err)
	}

	if n != int64(len(data)) {
		t.Errorf("expected %d bytes, got %d", len(data), n)
	}

	if !bytes.Equal(dst.Bytes(), data) {
		t.Error("copied data doesn't match")
	}
}

func TestZeroCopy_LargeTransfer(t *testing.T) {
	// Create 1 MB of data
	data := make([]byte, 1024*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	src := bytes.NewReader(data)
	dst := &bytes.Buffer{}

	n, err := network.CopyWithZeroCopy(dst, src)
	if err != nil {
		t.Fatalf("large copy failed: %v", err)
	}

	if n != int64(len(data)) {
		t.Errorf("expected %d bytes, got %d", len(data), n)
	}

	if !bytes.Equal(dst.Bytes(), data) {
		t.Error("large data copy mismatch")
	}
}

func TestZeroCopy_MultiCopy(t *testing.T) {
	pairs := []network.CopyPair{
		{Src: strings.NewReader("data1"), Dst: &bytes.Buffer{}},
		{Src: strings.NewReader("data2"), Dst: &bytes.Buffer{}},
		{Src: strings.NewReader("data3"), Dst: &bytes.Buffer{}},
	}

	err := network.MultiCopy(pairs)
	if err != nil {
		t.Fatalf("multi copy failed: %v", err)
	}

	// Verify each copy
	expected := []string{"data1", "data2", "data3"}
	for i, pair := range pairs {
		buf := pair.Dst.(*bytes.Buffer)
		if buf.String() != expected[i] {
			t.Errorf("pair %d: expected %s, got %s", i, expected[i], buf.String())
		}
	}
}

func TestBufferedWriter(t *testing.T) {
	dst := &bytes.Buffer{}
	bw := network.NewBufferedWriter(dst)
	defer bw.Close()

	data := []byte("buffered write test")
	n, err := bw.Write(data)
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}

	if n != len(data) {
		t.Errorf("expected %d bytes written, got %d", len(data), n)
	}

	// Flush to underlying writer
	err = bw.Flush()
	if err != nil {
		t.Fatalf("flush failed: %v", err)
	}

	if !bytes.Equal(dst.Bytes(), data) {
		t.Error("buffered data doesn't match")
	}
}

func TestBufferedWriter_MultipleWrites(t *testing.T) {
	dst := &bytes.Buffer{}
	bw := network.NewBufferedWriter(dst)
	defer bw.Close()

	writes := [][]byte{
		[]byte("first "),
		[]byte("second "),
		[]byte("third"),
	}

	for _, data := range writes {
		_, err := bw.Write(data)
		if err != nil {
			t.Fatalf("write failed: %v", err)
		}
	}

	err := bw.Flush()
	if err != nil {
		t.Fatalf("flush failed: %v", err)
	}

	expected := "first second third"
	if dst.String() != expected {
		t.Errorf("expected %s, got %s", expected, dst.String())
	}
}

func TestStreamCopier(t *testing.T) {
	copier := network.NewStreamCopier(4)
	defer copier.Close()

	// Submit jobs
	jobs := []network.CopyJob{
		{ID: "job1", Src: strings.NewReader("data1"), Dst: &bytes.Buffer{}},
		{ID: "job2", Src: strings.NewReader("data2"), Dst: &bytes.Buffer{}},
		{ID: "job3", Src: strings.NewReader("data3"), Dst: &bytes.Buffer{}},
	}

	for _, job := range jobs {
		copier.Submit(job)
	}

	// Collect results
	results := make(map[string]network.CopyResult)
	for i := 0; i < len(jobs); i++ {
		result := <-copier.Results()
		if result.Error != nil {
			t.Errorf("job %s failed: %v", result.ID, result.Error)
		}
		results[result.ID] = result
	}

	// Verify all jobs completed
	for _, job := range jobs {
		if _, ok := results[job.ID]; !ok {
			t.Errorf("job %s not found in results", job.ID)
		}
	}
}

func BenchmarkZeroCopy_SmallData(b *testing.B) {
	data := make([]byte, 1024) // 1 KB
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		src := bytes.NewReader(data)
		dst := &bytes.Buffer{}
		network.CopyWithZeroCopy(dst, src)
	}
}

func BenchmarkZeroCopy_LargeData(b *testing.B) {
	data := make([]byte, 1024*1024) // 1 MB
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		src := bytes.NewReader(data)
		dst := &bytes.Buffer{}
		network.CopyWithZeroCopy(dst, src)
	}
}

func BenchmarkZeroCopy_vs_StandardCopy(b *testing.B) {
	data := make([]byte, 64*1024) // 64 KB

	b.Run("ZeroCopy", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			src := bytes.NewReader(data)
			dst := &bytes.Buffer{}
			network.CopyWithZeroCopy(dst, src)
		}
	})

	b.Run("StandardCopy", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			src := bytes.NewReader(data)
			dst := &bytes.Buffer{}
			io.Copy(dst, src)
		}
	})
}

func BenchmarkBufferedWriter(b *testing.B) {
	data := make([]byte, 1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dst := &bytes.Buffer{}
		bw := network.NewBufferedWriter(dst)
		bw.Write(data)
		bw.Flush()
		bw.Close()
	}
}

func BenchmarkStreamCopier_Parallel(b *testing.B) {
	copier := network.NewStreamCopier(8)
	defer copier.Close()

	data := make([]byte, 4096)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		job := network.CopyJob{
			ID:  "bench",
			Src: bytes.NewReader(data),
			Dst: io.Discard,
		}
		copier.Submit(job)
		<-copier.Results()
	}
}
