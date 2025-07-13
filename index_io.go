package faiss

/*
#include <stdlib.h>
#include <faiss/c_api/index_io_c.h>
*/
import "C"
import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unsafe"
)

// IO flags for index reading/writing
const (
	IOFlagMmap     = C.FAISS_IO_FLAG_MMAP      // Memory-map the index file
	IOFlagReadOnly = C.FAISS_IO_FLAG_READ_ONLY // Open in read-only mode
)

// WriteIndex writes an index to a file.
// The index is serialized in the FAISS binary format.
func WriteIndex(idx Index, filename string) error {
	if idx == nil {
		return fmt.Errorf("index is nil")
	}

	if filename == "" {
		return fmt.Errorf("filename is empty")
	}

	// Validate the path
	if err := ValidateFilePath(filename, false); err != nil {
		return wrapError(err, "write index path validation")
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return wrapError(err, "create directory")
	}

	cfname := C.CString(filename)
	defer C.free(unsafe.Pointer(cfname))

	if c := C.faiss_write_index_fname(idx.cPtr(), cfname); c != 0 {
		return wrapError(getLastError(), "write index")
	}

	return nil
}

// ReadIndex reads an index from a file.
// The ioflags parameter controls how the file is opened.
func ReadIndex(filename string, ioflags int) (*IndexImpl, error) {
	if filename == "" {
		return nil, fmt.Errorf("filename is empty")
	}

	// Validate the path
	if err := ValidateFilePath(filename, true); err != nil {
		return nil, wrapError(err, "read index path validation")
	}

	cfname := C.CString(filename)
	defer C.free(unsafe.Pointer(cfname))

	var cIdx *C.FaissIndex
	if c := C.faiss_read_index_fname(cfname, C.int(ioflags), &cIdx); c != 0 {
		return nil, wrapError(getLastError(), "read index")
	}

	idx := &faissIndex{idx: cIdx}
	runtime.SetFinalizer(idx, (*faissIndex).Delete)

	return &IndexImpl{idx}, nil
}

// WriteIndexBinary writes an index to a file with binary format
func WriteIndexBinary(idx Index, filename string) error {
	return WriteIndex(idx, filename)
}

// ReadIndexBinary reads an index from a binary file
func ReadIndexBinary(filename string) (*IndexImpl, error) {
	return ReadIndex(filename, 0)
}

// WriteIndexMmap writes an index optimized for memory mapping
func WriteIndexMmap(idx Index, filename string) error {
	return WriteIndex(idx, filename)
}

// ReadIndexMmap reads an index with memory mapping
func ReadIndexMmap(filename string) (*IndexImpl, error) {
	return ReadIndex(filename, IOFlagMmap)
}

// WriteIndexReadOnly writes an index for read-only access
func WriteIndexReadOnly(idx Index, filename string) error {
	return WriteIndex(idx, filename)
}

// ReadIndexReadOnly reads an index in read-only mode
func ReadIndexReadOnly(filename string) (*IndexImpl, error) {
	return ReadIndex(filename, IOFlagReadOnly)
}

// IndexFileInfo contains information about an index file
type IndexFileInfo struct {
	Filename   string
	Size       int64
	ModTime    time.Time
	IsReadable bool
	IsWritable bool
	Exists     bool
	Extension  string
	Directory  string
	Basename   string
}

// GetIndexFileInfo returns information about an index file
func GetIndexFileInfo(filename string) (*IndexFileInfo, error) {
	if filename == "" {
		return nil, fmt.Errorf("filename is empty")
	}

	info := &IndexFileInfo{
		Filename:  filename,
		Extension: filepath.Ext(filename),
		Directory: filepath.Dir(filename),
		Basename:  filepath.Base(filename),
	}

	stat, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			info.Exists = false
			return info, nil
		}
		return nil, wrapError(err, "get file info")
	}

	info.Exists = true
	info.Size = stat.Size()
	info.ModTime = stat.ModTime()
	info.IsReadable = IsFileReadable(filename)
	info.IsWritable = IsFileWritable(filename)

	return info, nil
}

// ValidateFilePath validates a file path for index operations
func ValidateFilePath(filename string, mustExist bool) error {
	if filename == "" {
		return fmt.Errorf("filename is empty")
	}

	// Check for invalid characters
	if strings.ContainsAny(filename, "<>:\"|?*") {
		return fmt.Errorf("filename contains invalid characters")
	}

	// Check if file exists when required
	if mustExist {
		if _, err := os.Stat(filename); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("file does not exist: %s", filename)
			}
			return wrapError(err, "check file existence")
		}
	}

	return nil
}

// IsFileReadable checks if a file is readable
func IsFileReadable(filename string) bool {
	file, err := os.Open(filename)
	if err != nil {
		return false
	}
	defer file.Close()
	return true
}

// IsFileWritable checks if a file is writable
func IsFileWritable(filename string) bool {
	// Try to open for writing
	file, err := os.OpenFile(filename, os.O_WRONLY, 0)
	if err != nil {
		return false
	}
	defer file.Close()
	return true
}

// BackupIndex creates a backup of an index file
func BackupIndex(originalPath, backupPath string) error {
	if originalPath == "" || backupPath == "" {
		return fmt.Errorf("paths cannot be empty")
	}

	// Validate original file exists
	if err := ValidateFilePath(originalPath, true); err != nil {
		return wrapError(err, "validate original file")
	}

	// Create backup directory if needed
	backupDir := filepath.Dir(backupPath)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return wrapError(err, "create backup directory")
	}

	// Copy file
	if err := copyFile(originalPath, backupPath); err != nil {
		return wrapError(err, "copy file")
	}

	return nil
}

// RestoreIndex restores an index from a backup
func RestoreIndex(backupPath, targetPath string) error {
	if backupPath == "" || targetPath == "" {
		return fmt.Errorf("paths cannot be empty")
	}

	// Validate backup file exists
	if err := ValidateFilePath(backupPath, true); err != nil {
		return wrapError(err, "validate backup file")
	}

	// Create target directory if needed
	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return wrapError(err, "create target directory")
	}

	// Copy file
	if err := copyFile(backupPath, targetPath); err != nil {
		return wrapError(err, "copy file")
	}

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy file contents
	buffer := make([]byte, 64*1024) // 64KB buffer
	for {
		n, err := sourceFile.Read(buffer)
		if err != nil && err.Error() != "EOF" {
			return err
		}
		if n == 0 {
			break
		}

		if _, err := destFile.Write(buffer[:n]); err != nil {
			return err
		}
	}

	return destFile.Sync()
}

// IndexSerializer provides utilities for index serialization
type IndexSerializer struct {
	CompressionLevel int
	UseBackup        bool
	BackupSuffix     string
	Verbose          bool
}

// NewIndexSerializer creates a new index serializer
func NewIndexSerializer() *IndexSerializer {
	return &IndexSerializer{
		CompressionLevel: 0,
		UseBackup:        false,
		BackupSuffix:     ".bak",
		Verbose:          false,
	}
}

// SetCompressionLevel sets the compression level (placeholder)
func (s *IndexSerializer) SetCompressionLevel(level int) *IndexSerializer {
	s.CompressionLevel = level
	return s
}

// SetUseBackup enables/disables backup creation
func (s *IndexSerializer) SetUseBackup(useBackup bool) *IndexSerializer {
	s.UseBackup = useBackup
	return s
}

// SetBackupSuffix sets the backup file suffix
func (s *IndexSerializer) SetBackupSuffix(suffix string) *IndexSerializer {
	s.BackupSuffix = suffix
	return s
}

// SetVerbose enables/disables verbose logging
func (s *IndexSerializer) SetVerbose(verbose bool) *IndexSerializer {
	s.Verbose = verbose
	return s
}

// WriteIndex writes an index using the serializer settings
func (s *IndexSerializer) WriteIndex(idx Index, filename string) error {
	if s.Verbose {
		fmt.Printf("Writing index to %s\n", filename)
	}

	// Create backup if enabled
	if s.UseBackup {
		if _, err := os.Stat(filename); err == nil {
			backupPath := filename + s.BackupSuffix
			if err := BackupIndex(filename, backupPath); err != nil {
				return wrapError(err, "create backup")
			}
			if s.Verbose {
				fmt.Printf("Created backup at %s\n", backupPath)
			}
		}
	}

	// Write index
	if err := WriteIndex(idx, filename); err != nil {
		return err
	}

	if s.Verbose {
		info, _ := GetIndexFileInfo(filename)
		if info != nil {
			fmt.Printf("Index written successfully, size: %d bytes\n", info.Size)
		}
	}

	return nil
}

// ReadIndex reads an index using the serializer settings
func (s *IndexSerializer) ReadIndex(filename string) (*IndexImpl, error) {
	if s.Verbose {
		fmt.Printf("Reading index from %s\n", filename)
	}

	// Read index
	idx, err := ReadIndex(filename, 0)
	if err != nil {
		return nil, err
	}

	if s.Verbose {
		fmt.Printf("Index loaded successfully, dimension: %d, ntotal: %d\n",
			idx.D(), idx.Ntotal())
	}

	return idx, nil
}

// BatchIndexManager manages multiple index files
type BatchIndexManager struct {
	BasePath   string
	Serializer *IndexSerializer
	Indices    map[string]*IndexImpl
}

// NewBatchIndexManager creates a new batch index manager
func NewBatchIndexManager(basePath string) *BatchIndexManager {
	return &BatchIndexManager{
		BasePath:   basePath,
		Serializer: NewIndexSerializer(),
		Indices:    make(map[string]*IndexImpl),
	}
}

// AddIndex adds an index to the manager
func (m *BatchIndexManager) AddIndex(name string, idx *IndexImpl) {
	m.Indices[name] = idx
}

// SaveAll saves all indices to disk
func (m *BatchIndexManager) SaveAll() error {
	for name, idx := range m.Indices {
		filename := filepath.Join(m.BasePath, name+".faiss")
		if err := m.Serializer.WriteIndex(idx, filename); err != nil {
			return wrapError(err, fmt.Sprintf("save index %s", name))
		}
	}
	return nil
}

// LoadAll loads all indices from disk
func (m *BatchIndexManager) LoadAll() error {
	// Clear existing indices
	m.Indices = make(map[string]*IndexImpl)

	// Find all .faiss files in the base path
	files, err := filepath.Glob(filepath.Join(m.BasePath, "*.faiss"))
	if err != nil {
		return wrapError(err, "find index files")
	}

	for _, file := range files {
		name := strings.TrimSuffix(filepath.Base(file), ".faiss")
		idx, err := m.Serializer.ReadIndex(file)
		if err != nil {
			return wrapError(err, fmt.Sprintf("load index %s", name))
		}
		m.Indices[name] = idx
	}

	return nil
}

// GetIndex retrieves an index by name
func (m *BatchIndexManager) GetIndex(name string) (*IndexImpl, bool) {
	idx, exists := m.Indices[name]
	return idx, exists
}

// DeleteIndex removes an index from the manager
func (m *BatchIndexManager) DeleteIndex(name string) {
	if idx, exists := m.Indices[name]; exists {
		idx.Delete()
		delete(m.Indices, name)
	}
}

// ListIndices returns a list of all index names
func (m *BatchIndexManager) ListIndices() []string {
	names := make([]string, 0, len(m.Indices))
	for name := range m.Indices {
		names = append(names, name)
	}
	return names
}

// GetIndexInfo returns information about all indices
func (m *BatchIndexManager) GetIndexInfo() map[string]IndexStats {
	info := make(map[string]IndexStats)
	for name, idx := range m.Indices {
		info[name] = IndexStats{
			Name:      name,
			Dimension: idx.D(),
			Count:     idx.Ntotal(),
			IsTrained: idx.IsTrained(),
			Metric:    idx.MetricType(),
		}
	}
	return info
}

// IndexStats contains statistics about an index
type IndexStats struct {
	Name      string
	Dimension int
	Count     int64
	IsTrained bool
	Metric    int
}

// Close closes all indices and cleans up resources
func (m *BatchIndexManager) Close() {
	for _, idx := range m.Indices {
		idx.Delete()
	}
	m.Indices = nil
}
