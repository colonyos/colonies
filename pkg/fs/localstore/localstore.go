package localstore

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

)

const bufferSize = 32 * 1024

// LocalObjectStore is a filesystem-backed ObjectStore implementation.
//
// Storage layout:
//
//	{baseDir}/{colonyName}/objects/{objectName}
//	{baseDir}/{colonyName}/.staging/{objectName}/{chunkIndex}
type LocalObjectStore struct {
	mu      sync.RWMutex
	baseDir string
}

// NewLocalObjectStore creates a new LocalObjectStore rooted at baseDir.
func NewLocalObjectStore(baseDir string) (*LocalObjectStore, error) {
	absDir, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve base dir: %w", err)
	}

	if err := os.MkdirAll(absDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base dir: %w", err)
	}

	return &LocalObjectStore{baseDir: absDir}, nil
}

func (s *LocalObjectStore) validateName(name string) error {
	if name == "" {
		return fmt.Errorf("name must not be empty")
	}
	if strings.Contains(name, "..") {
		return fmt.Errorf("name must not contain '..'")
	}
	if strings.ContainsAny(name, "/\\") {
		return fmt.Errorf("name must not contain path separators")
	}
	if strings.ContainsRune(name, 0) {
		return fmt.Errorf("name must not contain null bytes")
	}
	return nil
}

func (s *LocalObjectStore) objectDir(colonyName string) string {
	return filepath.Join(s.baseDir, colonyName, "objects")
}

func (s *LocalObjectStore) objectPath(colonyName, objectName string) string {
	return filepath.Join(s.objectDir(colonyName), objectName)
}

func (s *LocalObjectStore) stagingDir(colonyName, objectName string) string {
	return filepath.Join(s.baseDir, colonyName, ".staging", objectName)
}

func (s *LocalObjectStore) verifyWithinBase(path string) error {
	resolved, err := filepath.EvalSymlinks(filepath.Dir(path))
	if err != nil {
		// Parent dir may not exist yet; verify the grandparent
		resolved, err = filepath.EvalSymlinks(filepath.Dir(filepath.Dir(path)))
		if err != nil {
			return fmt.Errorf("path verification failed: %w", err)
		}
	}
	if !strings.HasPrefix(resolved, s.baseDir) {
		return fmt.Errorf("path escapes base directory")
	}
	return nil
}

func (s *LocalObjectStore) Put(colonyName, objectName string, reader io.Reader, size int64) error {
	if err := s.validateName(colonyName); err != nil {
		return fmt.Errorf("invalid colony name: %w", err)
	}
	if err := s.validateName(objectName); err != nil {
		return fmt.Errorf("invalid object name: %w", err)
	}

	objDir := s.objectDir(colonyName)
	if err := os.MkdirAll(objDir, 0755); err != nil {
		return fmt.Errorf("failed to create object dir: %w", err)
	}

	target := s.objectPath(colonyName, objectName)
	if err := s.verifyWithinBase(target); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	return atomicWrite(objDir, target, reader)
}

func (s *LocalObjectStore) PutChunk(colonyName, objectName string, chunkIndex int, reader io.Reader, size int64) error {
	if err := s.validateName(colonyName); err != nil {
		return fmt.Errorf("invalid colony name: %w", err)
	}
	if err := s.validateName(objectName); err != nil {
		return fmt.Errorf("invalid object name: %w", err)
	}
	if chunkIndex < 0 {
		return fmt.Errorf("chunk index must be non-negative")
	}

	stagingDir := s.stagingDir(colonyName, objectName)
	if err := os.MkdirAll(stagingDir, 0755); err != nil {
		return fmt.Errorf("failed to create staging dir: %w", err)
	}

	target := filepath.Join(stagingDir, strconv.Itoa(chunkIndex))
	if err := s.verifyWithinBase(target); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	return atomicWrite(stagingDir, target, reader)
}

func (s *LocalObjectStore) AssembleChunks(colonyName, objectName string, totalChunks int) error {
	if err := s.validateName(colonyName); err != nil {
		return fmt.Errorf("invalid colony name: %w", err)
	}
	if err := s.validateName(objectName); err != nil {
		return fmt.Errorf("invalid object name: %w", err)
	}

	stagingDir := s.stagingDir(colonyName, objectName)
	objDir := s.objectDir(colonyName)
	if err := os.MkdirAll(objDir, 0755); err != nil {
		return fmt.Errorf("failed to create object dir: %w", err)
	}

	target := s.objectPath(colonyName, objectName)
	if err := s.verifyWithinBase(target); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Verify all chunks exist
	for i := 0; i < totalChunks; i++ {
		chunkPath := filepath.Join(stagingDir, strconv.Itoa(i))
		if _, err := os.Stat(chunkPath); err != nil {
			return fmt.Errorf("missing chunk %d: %w", i, err)
		}
	}

	// Create temp file and stream-append all chunks
	tmp, err := os.CreateTemp(objDir, ".tmp-assemble-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpName := tmp.Name()

	buf := make([]byte, bufferSize)
	for i := 0; i < totalChunks; i++ {
		chunkPath := filepath.Join(stagingDir, strconv.Itoa(i))
		f, err := os.Open(chunkPath)
		if err != nil {
			tmp.Close()
			os.Remove(tmpName)
			return fmt.Errorf("failed to open chunk %d: %w", i, err)
		}
		_, err = io.CopyBuffer(tmp, f, buf)
		f.Close()
		if err != nil {
			tmp.Close()
			os.Remove(tmpName)
			return fmt.Errorf("failed to copy chunk %d: %w", i, err)
		}
	}

	if err := tmp.Sync(); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("failed to sync assembled file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("failed to close assembled file: %w", err)
	}

	if err := os.Rename(tmpName, target); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("failed to rename assembled file: %w", err)
	}

	// Clean up staging
	os.RemoveAll(stagingDir)

	return nil
}

func (s *LocalObjectStore) GetChunkStatus(colonyName, objectName string) ([]int, error) {
	if err := s.validateName(colonyName); err != nil {
		return nil, fmt.Errorf("invalid colony name: %w", err)
	}
	if err := s.validateName(objectName); err != nil {
		return nil, fmt.Errorf("invalid object name: %w", err)
	}

	stagingDir := s.stagingDir(colonyName, objectName)

	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(stagingDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []int{}, nil
		}
		return nil, fmt.Errorf("failed to read staging dir: %w", err)
	}

	var indices []int
	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".tmp-") {
			continue
		}
		idx, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}
		indices = append(indices, idx)
	}
	sort.Ints(indices)

	return indices, nil
}

func (s *LocalObjectStore) Get(colonyName, objectName string) (io.ReadCloser, int64, error) {
	if err := s.validateName(colonyName); err != nil {
		return nil, 0, fmt.Errorf("invalid colony name: %w", err)
	}
	if err := s.validateName(objectName); err != nil {
		return nil, 0, fmt.Errorf("invalid object name: %w", err)
	}

	path := s.objectPath(colonyName, objectName)

	s.mu.RLock()
	defer s.mu.RUnlock()

	f, err := os.Open(path)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to open object: %w", err)
	}

	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, 0, fmt.Errorf("failed to stat object: %w", err)
	}

	return f, info.Size(), nil
}

func (s *LocalObjectStore) GetRange(colonyName, objectName string, offset, length int64) (io.ReadCloser, error) {
	if err := s.validateName(colonyName); err != nil {
		return nil, fmt.Errorf("invalid colony name: %w", err)
	}
	if err := s.validateName(objectName); err != nil {
		return nil, fmt.Errorf("invalid object name: %w", err)
	}

	path := s.objectPath(colonyName, objectName)

	s.mu.RLock()
	defer s.mu.RUnlock()

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open object: %w", err)
	}

	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("failed to stat object: %w", err)
	}

	if offset >= info.Size() {
		f.Close()
		return nil, fmt.Errorf("offset %d is beyond file size %d", offset, info.Size())
	}

	if _, err := f.Seek(offset, io.SeekStart); err != nil {
		f.Close()
		return nil, fmt.Errorf("failed to seek: %w", err)
	}

	// Clamp length to remaining bytes
	remaining := info.Size() - offset
	if length > remaining {
		length = remaining
	}

	return &limitedReadCloser{reader: io.LimitReader(f, length), closer: f}, nil
}

func (s *LocalObjectStore) Remove(colonyName, objectName string) error {
	if err := s.validateName(colonyName); err != nil {
		return fmt.Errorf("invalid colony name: %w", err)
	}
	if err := s.validateName(objectName); err != nil {
		return fmt.Errorf("invalid object name: %w", err)
	}

	path := s.objectPath(colonyName, objectName)

	s.mu.Lock()
	defer s.mu.Unlock()

	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove object: %w", err)
	}
	return nil
}

func (s *LocalObjectStore) RemoveStaging(colonyName, objectName string) error {
	if err := s.validateName(colonyName); err != nil {
		return fmt.Errorf("invalid colony name: %w", err)
	}
	if err := s.validateName(objectName); err != nil {
		return fmt.Errorf("invalid object name: %w", err)
	}

	stagingDir := s.stagingDir(colonyName, objectName)

	s.mu.Lock()
	defer s.mu.Unlock()

	err := os.RemoveAll(stagingDir)
	if err != nil {
		return fmt.Errorf("failed to remove staging dir: %w", err)
	}
	return nil
}

func (s *LocalObjectStore) Exists(colonyName, objectName string) bool {
	if err := s.validateName(colonyName); err != nil {
		return false
	}
	if err := s.validateName(objectName); err != nil {
		return false
	}

	path := s.objectPath(colonyName, objectName)

	s.mu.RLock()
	defer s.mu.RUnlock()

	_, err := os.Stat(path)
	return err == nil
}

func (s *LocalObjectStore) Stat(colonyName, objectName string) (ObjectInfo, error) {
	if err := s.validateName(colonyName); err != nil {
		return ObjectInfo{}, fmt.Errorf("invalid colony name: %w", err)
	}
	if err := s.validateName(objectName); err != nil {
		return ObjectInfo{}, fmt.Errorf("invalid object name: %w", err)
	}

	path := s.objectPath(colonyName, objectName)

	s.mu.RLock()
	defer s.mu.RUnlock()

	info, err := os.Stat(path)
	if err != nil {
		return ObjectInfo{}, fmt.Errorf("failed to stat object: %w", err)
	}
	return ObjectInfo{Size: info.Size()}, nil
}

func (s *LocalObjectStore) List(colonyName string) ([]string, error) {
	if err := s.validateName(colonyName); err != nil {
		return nil, fmt.Errorf("invalid colony name: %w", err)
	}

	objDir := s.objectDir(colonyName)

	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(objDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read object dir: %w", err)
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".tmp-") {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	return names, nil
}

func (s *LocalObjectStore) RemoveAll(colonyName string) error {
	if err := s.validateName(colonyName); err != nil {
		return fmt.Errorf("invalid colony name: %w", err)
	}

	colonyDir := filepath.Join(s.baseDir, colonyName)

	s.mu.Lock()
	defer s.mu.Unlock()

	err := os.RemoveAll(colonyDir)
	if err != nil {
		return fmt.Errorf("failed to remove colony dir: %w", err)
	}
	return nil
}

func (s *LocalObjectStore) DiskUsage(colonyName string) (int64, error) {
	if err := s.validateName(colonyName); err != nil {
		return 0, fmt.Errorf("invalid colony name: %w", err)
	}

	colonyDir := filepath.Join(s.baseDir, colonyName)

	s.mu.RLock()
	defer s.mu.RUnlock()

	var total int64
	err := filepath.Walk(colonyDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to calculate disk usage: %w", err)
	}
	return total, nil
}

// atomicWrite writes data from reader to target via a temp file + rename.
func atomicWrite(dir, target string, reader io.Reader) error {
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpName := tmp.Name()

	buf := make([]byte, bufferSize)
	if _, err := io.CopyBuffer(tmp, reader, buf); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("failed to write data: %w", err)
	}

	if err := tmp.Sync(); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("failed to sync: %w", err)
	}

	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	if err := os.Rename(tmpName, target); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("failed to rename to target: %w", err)
	}

	return nil
}

// limitedReadCloser wraps io.Reader with a separate closer.
type limitedReadCloser struct {
	reader io.Reader
	closer io.Closer
}

func (l *limitedReadCloser) Read(p []byte) (int, error) {
	return l.reader.Read(p)
}

func (l *limitedReadCloser) Close() error {
	return l.closer.Close()
}
