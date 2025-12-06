// Package network provides delta synchronization for efficient layer transfers
package network

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash"
	"io"

	"github.com/cespare/xxhash/v2"
	"github.com/opencontainers/go-digest"
)

// DeltaSync implements rsync-like algorithm for layer synchronization
// It identifies common chunks between source and destination and only transfers differences
type DeltaSync struct {
	// ChunkSize is the size of each chunk for rolling hash (default: 1MB)
	ChunkSize int

	// WindowSize is the size of the rolling hash window (default: 64KB)
	WindowSize int

	// UseXXHash enables XXH64 instead of Adler32 for better distribution
	UseXXHash bool

	// EnableCompression enables compression of delta patches
	EnableCompression bool

	// MaxMemory is the maximum memory to use for chunk signatures (default: 100MB)
	MaxMemory int64
}

// ChunkSignature represents the signature of a data chunk
type ChunkSignature struct {
	// Offset is the chunk offset in the source data
	Offset int64

	// Size is the chunk size
	Size int

	// WeakHash is the rolling hash (weak checksum)
	WeakHash uint64

	// StrongHash is the cryptographic hash (strong checksum)
	StrongHash [32]byte
}

// Delta represents a difference between source and destination
type Delta struct {
	// Type is the delta operation type (COPY or DATA)
	Type DeltaType

	// Offset is the offset in the source (for COPY operations)
	Offset int64

	// Size is the size of the operation
	Size int

	// Data is the literal data (for DATA operations)
	Data []byte
}

// DeltaType represents the type of delta operation
type DeltaType int

const (
	// DeltaTypeCopy indicates copying from source
	DeltaTypeCopy DeltaType = iota

	// DeltaTypeData indicates literal data
	DeltaTypeData
)

// SyncResult contains the result of a delta sync operation
type SyncResult struct {
	// TotalBytes is the total size of the data
	TotalBytes int64

	// TransferredBytes is the number of bytes actually transferred
	TransferredBytes int64

	// Savings is the bandwidth savings percentage
	Savings float64

	// ChunkCount is the number of chunks processed
	ChunkCount int

	// MatchedChunks is the number of chunks that matched
	MatchedChunks int
}

// NewDeltaSync creates a new delta sync instance
func NewDeltaSync(chunkSize int) *DeltaSync {
	if chunkSize <= 0 {
		chunkSize = 1024 * 1024 // Default 1MB
	}

	return &DeltaSync{
		ChunkSize:         chunkSize,
		WindowSize:        64 * 1024, // 64KB
		UseXXHash:         true,
		EnableCompression: true,
		MaxMemory:         100 * 1024 * 1024, // 100MB
	}
}

// Sync synchronizes data from source to destination using delta algorithm
func (d *DeltaSync) Sync(ctx context.Context, src, dst io.ReadSeeker, output io.Writer) (*SyncResult, error) {
	// Generate chunk signatures from destination
	dstSignatures, err := d.generateSignatures(ctx, dst)
	if err != nil {
		return nil, fmt.Errorf("failed to generate destination signatures: %w", err)
	}

	// Build signature lookup map for fast matching
	signatureMap := d.buildSignatureMap(dstSignatures)

	// Reset source to beginning
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to reset source: %w", err)
	}

	// Generate and apply delta
	result, err := d.generateDelta(ctx, src, signatureMap, output)
	if err != nil {
		return nil, fmt.Errorf("failed to generate delta: %w", err)
	}

	return result, nil
}

// SyncLayer synchronizes a container layer with delta algorithm
func (d *DeltaSync) SyncLayer(ctx context.Context, srcReader, dstReader io.ReadSeeker, output io.Writer) (*SyncResult, error) {
	return d.Sync(ctx, srcReader, dstReader, output)
}

// generateSignatures generates chunk signatures from a data stream
func (d *DeltaSync) generateSignatures(ctx context.Context, reader io.ReadSeeker) ([]ChunkSignature, error) {
	var signatures []ChunkSignature
	chunk := make([]byte, d.ChunkSize)
	offset := int64(0)

	for {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Read chunk
		n, err := io.ReadFull(reader, chunk)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				// Process final chunk if any
				if n > 0 {
					sig := d.computeSignature(chunk[:n], offset)
					signatures = append(signatures, sig)
				}
				break
			}
			return nil, fmt.Errorf("failed to read chunk: %w", err)
		}

		// Compute signature for chunk
		sig := d.computeSignature(chunk[:n], offset)
		signatures = append(signatures, sig)

		offset += int64(n)
	}

	return signatures, nil
}

// computeSignature computes weak and strong signatures for a chunk
func (d *DeltaSync) computeSignature(data []byte, offset int64) ChunkSignature {
	sig := ChunkSignature{
		Offset: offset,
		Size:   len(data),
	}

	// Compute weak hash (rolling hash)
	if d.UseXXHash {
		sig.WeakHash = xxhash.Sum64(data)
	} else {
		sig.WeakHash = adler32Hash(data)
	}

	// Compute strong hash (cryptographic hash)
	sig.StrongHash = sha256.Sum256(data)

	return sig
}

// buildSignatureMap builds a map for fast signature lookup
func (d *DeltaSync) buildSignatureMap(signatures []ChunkSignature) map[uint64][]ChunkSignature {
	sigMap := make(map[uint64][]ChunkSignature)

	for _, sig := range signatures {
		sigMap[sig.WeakHash] = append(sigMap[sig.WeakHash], sig)
	}

	return sigMap
}

// generateDelta generates delta operations by comparing source with signatures
func (d *DeltaSync) generateDelta(ctx context.Context, src io.ReadSeeker, sigMap map[uint64][]ChunkSignature, output io.Writer) (*SyncResult, error) {
	result := &SyncResult{}

	// Rolling window buffer
	window := make([]byte, d.WindowSize)
	_ = make([]byte, d.ChunkSize) // Reserved for future use

	// Read initial window
	n, err := io.ReadFull(src, window)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, fmt.Errorf("failed to read initial window: %w", err)
	}

	offset := int64(0)
	windowData := window[:n]
	var literalBuffer []byte

	for len(windowData) > 0 {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Compute rolling hash for current window
		var weakHash uint64
		if d.UseXXHash {
			weakHash = xxhash.Sum64(windowData)
		} else {
			weakHash = adler32Hash(windowData)
		}

		// Check if this window matches any signature
		matched := false
		if candidates, ok := sigMap[weakHash]; ok {
			// Verify with strong hash
			strongHash := sha256.Sum256(windowData)

			for _, candidate := range candidates {
				if candidate.StrongHash == strongHash && candidate.Size == len(windowData) {
					// Match found - emit any pending literal data first
					if len(literalBuffer) > 0 {
						delta := Delta{
							Type: DeltaTypeData,
							Size: len(literalBuffer),
							Data: literalBuffer,
						}
						if err := d.writeDelta(output, delta); err != nil {
							return nil, err
						}
						result.TransferredBytes += int64(len(literalBuffer))
						literalBuffer = nil
					}

					// Emit copy operation
					delta := Delta{
						Type:   DeltaTypeCopy,
						Offset: candidate.Offset,
						Size:   candidate.Size,
					}
					if err := d.writeDelta(output, delta); err != nil {
						return nil, err
					}

					result.MatchedChunks++
					result.ChunkCount++

					// Skip ahead past the matched chunk
					skipSize := len(windowData)
					if _, err := src.Seek(int64(skipSize), io.SeekCurrent); err != nil {
						return nil, fmt.Errorf("failed to seek: %w", err)
					}
					offset += int64(skipSize)

					// Read next window
					n, err := io.ReadFull(src, window)
					if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
						return nil, fmt.Errorf("failed to read window: %w", err)
					}
					windowData = window[:n]
					matched = true
					break
				}
			}
		}

		if !matched {
			// No match - add first byte to literal buffer and slide window
			literalBuffer = append(literalBuffer, windowData[0])

			// Read one more byte and slide window
			var nextByte [1]byte
			_, err := io.ReadFull(src, nextByte[:])
			if err != nil {
				if err == io.EOF || err == io.ErrUnexpectedEOF {
					// End of stream - emit remaining data
					if len(windowData) > 1 {
						literalBuffer = append(literalBuffer, windowData[1:]...)
					}
					windowData = nil
				} else {
					return nil, fmt.Errorf("failed to read next byte: %w", err)
				}
			} else {
				// Slide window
				copy(windowData, windowData[1:])
				windowData[len(windowData)-1] = nextByte[0]
			}
			offset++
		}
	}

	// Emit any remaining literal data
	if len(literalBuffer) > 0 {
		delta := Delta{
			Type: DeltaTypeData,
			Size: len(literalBuffer),
			Data: literalBuffer,
		}
		if err := d.writeDelta(output, delta); err != nil {
			return nil, err
		}
		result.TransferredBytes += int64(len(literalBuffer))
	}

	// Calculate savings
	result.TotalBytes = offset
	if result.TotalBytes > 0 {
		result.Savings = float64(result.TotalBytes-result.TransferredBytes) / float64(result.TotalBytes) * 100
	}

	return result, nil
}

// writeDelta writes a delta operation to the output
func (d *DeltaSync) writeDelta(output io.Writer, delta Delta) error {
	// Write delta type (1 byte)
	if err := binary.Write(output, binary.LittleEndian, byte(delta.Type)); err != nil {
		return fmt.Errorf("failed to write delta type: %w", err)
	}

	// Write delta size (4 bytes)
	if err := binary.Write(output, binary.LittleEndian, uint32(delta.Size)); err != nil {
		return fmt.Errorf("failed to write delta size: %w", err)
	}

	switch delta.Type {
	case DeltaTypeCopy:
		// Write source offset (8 bytes)
		if err := binary.Write(output, binary.LittleEndian, uint64(delta.Offset)); err != nil {
			return fmt.Errorf("failed to write offset: %w", err)
		}

	case DeltaTypeData:
		// Write literal data
		if _, err := output.Write(delta.Data); err != nil {
			return fmt.Errorf("failed to write data: %w", err)
		}
	}

	return nil
}

// ApplyDelta applies a delta stream to reconstruct the target
func (d *DeltaSync) ApplyDelta(ctx context.Context, base io.ReadSeeker, delta io.Reader, output io.Writer) error {
	for {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Read delta type
		var deltaType byte
		if err := binary.Read(delta, binary.LittleEndian, &deltaType); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read delta type: %w", err)
		}

		// Read delta size
		var size uint32
		if err := binary.Read(delta, binary.LittleEndian, &size); err != nil {
			return fmt.Errorf("failed to read delta size: %w", err)
		}

		switch DeltaType(deltaType) {
		case DeltaTypeCopy:
			// Read source offset
			var offset uint64
			if err := binary.Read(delta, binary.LittleEndian, &offset); err != nil {
				return fmt.Errorf("failed to read offset: %w", err)
			}

			// Copy from base
			if _, err := base.Seek(int64(offset), io.SeekStart); err != nil {
				return fmt.Errorf("failed to seek base: %w", err)
			}

			if _, err := io.CopyN(output, base, int64(size)); err != nil {
				return fmt.Errorf("failed to copy from base: %w", err)
			}

		case DeltaTypeData:
			// Copy literal data
			if _, err := io.CopyN(output, delta, int64(size)); err != nil {
				return fmt.Errorf("failed to copy literal data: %w", err)
			}
		}
	}

	return nil
}

// EstimateSavings estimates potential bandwidth savings for a layer
func (d *DeltaSync) EstimateSavings(ctx context.Context, src, dst io.ReadSeeker) (float64, error) {
	// Generate signatures from destination
	dstSignatures, err := d.generateSignatures(ctx, dst)
	if err != nil {
		return 0, fmt.Errorf("failed to generate signatures: %w", err)
	}

	// Build signature map
	sigMap := d.buildSignatureMap(dstSignatures)

	// Reset source
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return 0, fmt.Errorf("failed to reset source: %w", err)
	}

	// Count matches
	chunk := make([]byte, d.ChunkSize)
	offset := int64(0)
	totalBytes := int64(0)
	matchedBytes := int64(0)

	for {
		n, err := io.ReadFull(src, chunk)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				if n > 0 {
					totalBytes += int64(n)
				}
				break
			}
			return 0, fmt.Errorf("failed to read chunk: %w", err)
		}

		totalBytes += int64(n)

		// Check if chunk matches
		sig := d.computeSignature(chunk[:n], offset)
		if candidates, ok := sigMap[sig.WeakHash]; ok {
			for _, candidate := range candidates {
				if candidate.StrongHash == sig.StrongHash {
					matchedBytes += int64(n)
					break
				}
			}
		}

		offset += int64(n)
	}

	if totalBytes == 0 {
		return 0, nil
	}

	savings := float64(matchedBytes) / float64(totalBytes) * 100
	return savings, nil
}

// adler32Hash computes Adler-32 rolling hash
func adler32Hash(data []byte) uint64 {
	const mod = 65521
	var a, b uint32 = 1, 0

	for _, byte := range data {
		a = (a + uint32(byte)) % mod
		b = (b + a) % mod
	}

	return uint64(b<<16 | a)
}

// VerifyDelta verifies a delta by reconstructing and comparing with source
func (d *DeltaSync) VerifyDelta(ctx context.Context, base io.ReadSeeker, delta io.Reader, expectedDigest digest.Digest) error {
	// Create a hash writer
	h := sha256.New()

	// Apply delta to reconstruction writer
	if err := d.ApplyDelta(ctx, base, delta, h); err != nil {
		return fmt.Errorf("failed to apply delta: %w", err)
	}

	// Verify digest
	actualDigest := digest.NewDigest(digest.SHA256, h)
	if actualDigest != expectedDigest {
		return fmt.Errorf("digest mismatch: expected %s, got %s", expectedDigest, actualDigest)
	}

	return nil
}

// hashWriter is an io.Writer that computes a hash
type hashWriter struct {
	hash hash.Hash
}

func (h *hashWriter) Write(p []byte) (n int, err error) {
	return h.hash.Write(p)
}
