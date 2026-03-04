package wal

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"sync"
)

var crcTable = crc32.MakeTable(crc32.IEEE)


// Binary format for each entry:
//
//	[total_len:4][crc32:4][op:1][entity_len:2][entity][key_len:2][key][ts:8][data_len:4][data]
//
// total_len covers everything after itself (crc32 through data).
// crc32 covers everything after itself (op through data).
const (
	headerSize  = 4 + 4 // total_len + crc32
	fixedFields = 1 + 2 + 2 + 8 + 4
)

// FileWAL is a WAL implementation backed by a single file on disk.
type FileWAL struct {
	mu       sync.Mutex
	file     *os.File
	bw       *bufio.Writer
	dir      string
	syncMode SyncMode
	encBuf   []byte // reusable encode buffer (only used under mu)
}

// NewFileWAL opens or creates a WAL file in the given directory.
func NewFileWAL(dir string, syncMode SyncMode) (*FileWAL, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create WAL directory: %w", err)
	}

	path := filepath.Join(dir, "wal.log")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open WAL file: %w", err)
	}

	return &FileWAL{
		file:     f,
		bw:       bufio.NewWriterSize(f, 64*1024),
		dir:      dir,
		syncMode: syncMode,
		encBuf:   make([]byte, 0, 1024),
	}, nil
}

// Append writes an entry to the WAL.
func (w *FileWAL) Append(entry Entry) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.encBuf = encodeEntryBuf(w.encBuf[:0], entry)
	if _, err := w.bw.Write(w.encBuf); err != nil {
		return fmt.Errorf("failed to write WAL entry: %w", err)
	}

	if w.syncMode == SyncAlways {
		if err := w.bw.Flush(); err != nil {
			return fmt.Errorf("failed to flush WAL buffer: %w", err)
		}
		if err := w.file.Sync(); err != nil {
			return fmt.Errorf("failed to sync WAL: %w", err)
		}
	}

	return nil
}

// Replay reads all valid entries from the WAL, calling fn for each.
// Corrupt or truncated trailing entries are silently skipped (crash recovery).
func (w *FileWAL) Replay(fn func(Entry) error) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Flush any buffered writes before replaying
	if err := w.bw.Flush(); err != nil {
		return fmt.Errorf("failed to flush WAL buffer before replay: %w", err)
	}

	path := filepath.Join(w.dir, "wal.log")
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to open WAL for replay: %w", err)
	}
	defer f.Close()

	br := bufio.NewReaderSize(f, 64*1024)
	for {
		entry, err := readEntry(br)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return nil // end of log or truncated entry
			}
			if _, ok := err.(*corruptEntryError); ok {
				return nil // corrupt trailing entry, stop replay
			}
			return err
		}

		if err := fn(entry); err != nil {
			return err
		}
	}
}

// Truncate removes all entries by replacing the WAL file with an empty one.
func (w *FileWAL) Truncate() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.bw.Flush(); err != nil {
		return fmt.Errorf("failed to flush WAL buffer for truncate: %w", err)
	}
	if err := w.file.Close(); err != nil {
		return fmt.Errorf("failed to close WAL for truncate: %w", err)
	}

	path := filepath.Join(w.dir, "wal.log")
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to truncate WAL: %w", err)
	}
	w.file = f
	w.bw.Reset(f)

	return nil
}

// Close closes the WAL file.
func (w *FileWAL) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file != nil {
		if err := w.bw.Flush(); err != nil {
			w.file.Close()
			w.file = nil
			return fmt.Errorf("failed to flush WAL buffer on close: %w", err)
		}
		if w.syncMode == SyncAlways {
			w.file.Sync()
		}
		err := w.file.Close()
		w.file = nil
		return err
	}
	return nil
}

type corruptEntryError struct {
	msg string
}

func (e *corruptEntryError) Error() string {
	return e.msg
}

// encodeEntryBuf appends the encoded entry to dst and returns the extended buffer.
// This avoids allocations when dst has sufficient capacity.
func encodeEntryBuf(dst []byte, e Entry) []byte {
	payloadSize := fixedFields + len(e.Entity) + len(e.Key) + len(e.Data)
	totalLen := 4 + payloadSize // crc32 + payload
	needed := 4 + totalLen      // total_len header + body

	// Grow dst if needed
	if cap(dst) < needed {
		dst = make([]byte, needed)
	} else {
		dst = dst[:needed]
	}

	// total_len
	binary.LittleEndian.PutUint32(dst[0:4], uint32(totalLen))

	// build payload (after crc32 placeholder)
	offset := 8 // skip total_len(4) + crc32(4)

	dst[offset] = byte(e.Op)
	offset++

	binary.LittleEndian.PutUint16(dst[offset:offset+2], uint16(len(e.Entity)))
	offset += 2
	copy(dst[offset:], e.Entity)
	offset += len(e.Entity)

	binary.LittleEndian.PutUint16(dst[offset:offset+2], uint16(len(e.Key)))
	offset += 2
	copy(dst[offset:], e.Key)
	offset += len(e.Key)

	binary.LittleEndian.PutUint64(dst[offset:offset+8], uint64(e.Timestamp))
	offset += 8

	binary.LittleEndian.PutUint32(dst[offset:offset+4], uint32(len(e.Data)))
	offset += 4
	copy(dst[offset:], e.Data)

	// compute crc32 over payload (everything after crc32 field)
	crc := crc32.Update(0, crcTable, dst[8:])
	binary.LittleEndian.PutUint32(dst[4:8], crc)

	return dst
}

// encodeEntry allocates and returns a new buffer with the encoded entry.
// Used in tests and where a standalone buffer is needed.
func encodeEntry(e Entry) []byte {
	return encodeEntryBuf(nil, e)
}

func readEntry(r io.Reader) (Entry, error) {
	var zero Entry

	// Read total_len
	var totalLenBuf [4]byte
	if _, err := io.ReadFull(r, totalLenBuf[:]); err != nil {
		return zero, err
	}
	totalLen := binary.LittleEndian.Uint32(totalLenBuf[:])

	if totalLen < 4+fixedFields || totalLen > 64*1024*1024 {
		return zero, &corruptEntryError{msg: fmt.Sprintf("invalid entry length: %d", totalLen)}
	}

	// Read the rest of the entry
	body := make([]byte, totalLen)
	if _, err := io.ReadFull(r, body); err != nil {
		return zero, err
	}

	// Verify CRC
	storedCRC := binary.LittleEndian.Uint32(body[0:4])
	computedCRC := crc32.Update(0, crcTable, body[4:])
	if storedCRC != computedCRC {
		return zero, &corruptEntryError{msg: "CRC mismatch"}
	}

	// Parse payload
	offset := 4 // skip crc32

	op := OpType(body[offset])
	offset++

	entityLen := int(binary.LittleEndian.Uint16(body[offset : offset+2]))
	offset += 2
	if offset+entityLen > len(body) {
		return zero, &corruptEntryError{msg: "entity length exceeds entry"}
	}
	entity := string(body[offset : offset+entityLen])
	offset += entityLen

	keyLen := int(binary.LittleEndian.Uint16(body[offset : offset+2]))
	offset += 2
	if offset+keyLen > len(body) {
		return zero, &corruptEntryError{msg: "key length exceeds entry"}
	}
	key := string(body[offset : offset+keyLen])
	offset += keyLen

	if offset+8 > len(body) {
		return zero, &corruptEntryError{msg: "timestamp exceeds entry"}
	}
	timestamp := int64(binary.LittleEndian.Uint64(body[offset : offset+8]))
	offset += 8

	if offset+4 > len(body) {
		return zero, &corruptEntryError{msg: "data length field exceeds entry"}
	}
	dataLen := int(binary.LittleEndian.Uint32(body[offset : offset+4]))
	offset += 4

	if offset+dataLen > len(body) {
		return zero, &corruptEntryError{msg: "data exceeds entry"}
	}

	var data []byte
	if dataLen > 0 {
		data = make([]byte, dataLen)
		copy(data, body[offset:offset+dataLen])
	}

	return Entry{
		Op:        op,
		Entity:    entity,
		Key:       key,
		Data:      data,
		Timestamp: timestamp,
	}, nil
}
