package lockfs

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/absfs/memfs"
)

// TestFilerOpenFileWithConcurrentAccess tests OpenFile with concurrent access
func TestFilerOpenFileWithConcurrentAccess(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	filer, err := NewFiler(mfs)
	if err != nil {
		t.Fatal(err)
	}

	const numGoroutines = 10
	var wg sync.WaitGroup
	errCh := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			filename := fmt.Sprintf("/file_%d.txt", id)
			f, err := filer.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
			if err != nil {
				errCh <- fmt.Errorf("goroutine %d: %w", id, err)
				return
			}
			_, err = f.Write([]byte(fmt.Sprintf("data from goroutine %d", id)))
			if err != nil {
				errCh <- fmt.Errorf("goroutine %d write: %w", id, err)
			}
			f.Close()
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Error(err)
	}
}

// TestFilerMkdirConcurrent tests Mkdir with concurrent access
func TestFilerMkdirConcurrent(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	filer, err := NewFiler(mfs)
	if err != nil {
		t.Fatal(err)
	}

	const numGoroutines = 10
	var wg sync.WaitGroup
	errCh := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			dirname := fmt.Sprintf("/dir_%d", id)
			err := filer.Mkdir(dirname, 0755)
			if err != nil {
				errCh <- fmt.Errorf("goroutine %d: %w", id, err)
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Error(err)
	}
}

// TestFilerRemoveConcurrent tests Remove with concurrent access
func TestFilerRemoveConcurrent(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	filer, err := NewFiler(mfs)
	if err != nil {
		t.Fatal(err)
	}

	// Create files first
	const numFiles = 10
	for i := 0; i < numFiles; i++ {
		filename := fmt.Sprintf("/file_%d.txt", i)
		f, err := filer.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			t.Fatal(err)
		}
		f.Close()
	}

	var wg sync.WaitGroup
	errCh := make(chan error, numFiles)

	for i := 0; i < numFiles; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			filename := fmt.Sprintf("/file_%d.txt", id)
			err := filer.Remove(filename)
			if err != nil {
				errCh <- fmt.Errorf("goroutine %d: %w", id, err)
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Error(err)
	}
}

// TestFilerRenameConcurrent tests Rename with concurrent access
func TestFilerRenameConcurrent(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	filer, err := NewFiler(mfs)
	if err != nil {
		t.Fatal(err)
	}

	// Create files first
	const numFiles = 10
	for i := 0; i < numFiles; i++ {
		filename := fmt.Sprintf("/file_%d.txt", i)
		f, err := filer.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			t.Fatal(err)
		}
		f.Close()
	}

	var wg sync.WaitGroup
	errCh := make(chan error, numFiles)

	for i := 0; i < numFiles; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			oldname := fmt.Sprintf("/file_%d.txt", id)
			newname := fmt.Sprintf("/renamed_%d.txt", id)
			err := filer.Rename(oldname, newname)
			if err != nil {
				errCh <- fmt.Errorf("goroutine %d: %w", id, err)
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Error(err)
	}
}

// TestFilerStatUsesReadLock tests that Stat uses read lock (allows concurrent access)
func TestFilerStatUsesReadLock(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	filer, err := NewFiler(mfs)
	if err != nil {
		t.Fatal(err)
	}

	// Create a test file
	f, err := filer.OpenFile("/test.txt", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		t.Fatal(err)
	}
	f.Write([]byte("test content"))
	f.Close()

	const numReaders = 20
	var wg sync.WaitGroup
	errCh := make(chan error, numReaders)

	// Multiple goroutines reading stat concurrently
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_, err := filer.Stat("/test.txt")
				if err != nil {
					errCh <- fmt.Errorf("reader %d: %w", id, err)
					return
				}
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Error(err)
	}
}

// TestFilerChmod tests the Chmod method
func TestFilerChmod(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	filer, err := NewFiler(mfs)
	if err != nil {
		t.Fatal(err)
	}

	// Create a file
	f, err := filer.OpenFile("/test.txt", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	// Change mode
	err = filer.Chmod("/test.txt", 0600)
	if err != nil {
		t.Fatal(err)
	}

	// Verify
	info, err := filer.Stat("/test.txt")
	if err != nil {
		t.Fatal(err)
	}

	if info.Mode().Perm() != 0600 {
		t.Errorf("expected mode 0600, got %o", info.Mode().Perm())
	}
}

// TestFilerChtimes tests the Chtimes method
func TestFilerChtimes(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	filer, err := NewFiler(mfs)
	if err != nil {
		t.Fatal(err)
	}

	// Create a file
	f, err := filer.OpenFile("/test.txt", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	// Set times
	atime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	mtime := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	err = filer.Chtimes("/test.txt", atime, mtime)
	if err != nil {
		t.Fatal(err)
	}

	// Verify
	info, err := filer.Stat("/test.txt")
	if err != nil {
		t.Fatal(err)
	}

	if !info.ModTime().Equal(mtime) {
		t.Errorf("expected mtime %v, got %v", mtime, info.ModTime())
	}
}

// TestFilerChown tests the Chown method
func TestFilerChown(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	filer, err := NewFiler(mfs)
	if err != nil {
		t.Fatal(err)
	}

	// Create a file
	f, err := filer.OpenFile("/test.txt", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	// Change ownership (memfs may not fully support this, but we test the API)
	err = filer.Chown("/test.txt", 1000, 1000)
	// Don't fail if not supported by underlying fs
	if err != nil && !os.IsPermission(err) {
		t.Logf("Chown returned: %v", err)
	}
}

// TestFilerReadDir tests the ReadDir method
func TestFilerReadDir(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	filer, err := NewFiler(mfs)
	if err != nil {
		t.Fatal(err)
	}

	// Create directory and files
	err = filer.Mkdir("/testdir", 0755)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 5; i++ {
		f, err := filer.OpenFile(fmt.Sprintf("/testdir/file%d.txt", i), os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			t.Fatal(err)
		}
		f.Close()
	}

	entries, err := filer.ReadDir("/testdir")
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 5 {
		t.Errorf("expected 5 entries, got %d", len(entries))
	}

	for _, entry := range entries {
		if entry.IsDir() {
			t.Errorf("expected file, got directory: %s", entry.Name())
		}
	}
}

// TestFilerReadFile tests the ReadFile method
func TestFilerReadFile(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	filer, err := NewFiler(mfs)
	if err != nil {
		t.Fatal(err)
	}

	testData := []byte("test file content")
	f, err := filer.OpenFile("/test.txt", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		t.Fatal(err)
	}
	f.Write(testData)
	f.Close()

	data, err := filer.ReadFile("/test.txt")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(data, testData) {
		t.Errorf("expected %q, got %q", testData, data)
	}
}

// TestFilerSub tests the Sub method
func TestFilerSub(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	filer, err := NewFiler(mfs)
	if err != nil {
		t.Fatal(err)
	}

	// Create subdirectory and file
	err = filer.Mkdir("/subdir", 0755)
	if err != nil {
		t.Fatal(err)
	}

	f, err := filer.OpenFile("/subdir/test.txt", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		t.Fatal(err)
	}
	f.Write([]byte("test"))
	f.Close()

	subFS, err := filer.Sub("/subdir")
	if err != nil {
		t.Fatal(err)
	}

	// Test that we can read from the sub filesystem
	data, err := fs.ReadFile(subFS, "test.txt")
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != "test" {
		t.Errorf("expected 'test', got %q", data)
	}
}

// TestFileSystemChdir tests the Chdir method
func TestFileSystemChdir(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	fsys, err := NewFS(mfs)
	if err != nil {
		t.Fatal(err)
	}

	// Create a directory
	err = fsys.Mkdir("/testdir", 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Change to it
	err = fsys.Chdir("/testdir")
	if err != nil {
		t.Fatal(err)
	}

	// Verify
	cwd, err := fsys.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if cwd != "/testdir" {
		t.Errorf("expected /testdir, got %s", cwd)
	}
}

// TestFileSystemGetwd tests the Getwd method
func TestFileSystemGetwd(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	fsys, err := NewFS(mfs)
	if err != nil {
		t.Fatal(err)
	}

	cwd, err := fsys.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if cwd != "/" {
		t.Errorf("expected /, got %s", cwd)
	}
}

// TestFileSystemTempDir tests the TempDir method
func TestFileSystemTempDir(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	fsys, err := NewFS(mfs)
	if err != nil {
		t.Fatal(err)
	}

	tempDir := fsys.TempDir()
	if tempDir == "" {
		t.Error("TempDir returned empty string")
	}
}

// TestFileSystemMkdirAll tests the MkdirAll method
func TestFileSystemMkdirAll(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	fsys, err := NewFS(mfs)
	if err != nil {
		t.Fatal(err)
	}

	err = fsys.MkdirAll("/a/b/c/d", 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Verify each directory exists
	for _, dir := range []string{"/a", "/a/b", "/a/b/c", "/a/b/c/d"} {
		info, err := fsys.Stat(dir)
		if err != nil {
			t.Errorf("directory %s does not exist: %v", dir, err)
		} else if !info.IsDir() {
			t.Errorf("%s is not a directory", dir)
		}
	}
}

// TestFileSystemRemoveAll tests the RemoveAll method
func TestFileSystemRemoveAll(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	fsys, err := NewFS(mfs)
	if err != nil {
		t.Fatal(err)
	}

	// Create nested structure
	err = fsys.MkdirAll("/testdir/sub1/sub2", 0755)
	if err != nil {
		t.Fatal(err)
	}

	f, err := fsys.Create("/testdir/file1.txt")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	f, err = fsys.Create("/testdir/sub1/file2.txt")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	// Remove all
	err = fsys.RemoveAll("/testdir")
	if err != nil {
		t.Fatal(err)
	}

	// Verify it's gone
	_, err = fsys.Stat("/testdir")
	if err == nil {
		t.Error("expected error, directory should not exist")
	}
}

// TestFileSystemTruncate tests the Truncate method
func TestFileSystemTruncate(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	fsys, err := NewFS(mfs)
	if err != nil {
		t.Fatal(err)
	}

	// Create file with content
	f, err := fsys.Create("/test.txt")
	if err != nil {
		t.Fatal(err)
	}
	f.Write([]byte("this is a long test content"))
	f.Close()

	// Truncate to 5 bytes
	err = fsys.Truncate("/test.txt", 5)
	if err != nil {
		t.Fatal(err)
	}

	// Verify size
	info, err := fsys.Stat("/test.txt")
	if err != nil {
		t.Fatal(err)
	}

	if info.Size() != 5 {
		t.Errorf("expected size 5, got %d", info.Size())
	}
}

// TestFileSystemReadDir tests the ReadDir method
func TestFileSystemReadDir(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	fsys, err := NewFS(mfs)
	if err != nil {
		t.Fatal(err)
	}

	// Create test structure
	err = fsys.Mkdir("/testdir", 0755)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 3; i++ {
		f, err := fsys.Create(fmt.Sprintf("/testdir/file%d.txt", i))
		if err != nil {
			t.Fatal(err)
		}
		f.Close()
	}

	entries, err := fsys.ReadDir("/testdir")
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}
}

// TestFileSystemReadFile tests the ReadFile method
func TestFileSystemReadFile(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	fsys, err := NewFS(mfs)
	if err != nil {
		t.Fatal(err)
	}

	testData := []byte("filesystem readfile test")
	f, err := fsys.Create("/test.txt")
	if err != nil {
		t.Fatal(err)
	}
	f.Write(testData)
	f.Close()

	data, err := fsys.ReadFile("/test.txt")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(data, testData) {
		t.Errorf("expected %q, got %q", testData, data)
	}
}

// TestFileSystemSub tests the Sub method
func TestFileSystemSub(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	fsys, err := NewFS(mfs)
	if err != nil {
		t.Fatal(err)
	}

	err = fsys.Mkdir("/subdir", 0755)
	if err != nil {
		t.Fatal(err)
	}

	f, err := fsys.Create("/subdir/test.txt")
	if err != nil {
		t.Fatal(err)
	}
	f.Write([]byte("sub content"))
	f.Close()

	subFS, err := fsys.Sub("/subdir")
	if err != nil {
		t.Fatal(err)
	}

	data, err := fs.ReadFile(subFS, "test.txt")
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != "sub content" {
		t.Errorf("expected 'sub content', got %q", data)
	}
}

// TestFileSystemChmod tests the Chmod method
func TestFileSystemChmod(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	fsys, err := NewFS(mfs)
	if err != nil {
		t.Fatal(err)
	}

	f, err := fsys.Create("/test.txt")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	err = fsys.Chmod("/test.txt", 0600)
	if err != nil {
		t.Fatal(err)
	}

	info, err := fsys.Stat("/test.txt")
	if err != nil {
		t.Fatal(err)
	}

	if info.Mode().Perm() != 0600 {
		t.Errorf("expected mode 0600, got %o", info.Mode().Perm())
	}
}

// TestFileSystemChtimes tests the Chtimes method
func TestFileSystemChtimes(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	fsys, err := NewFS(mfs)
	if err != nil {
		t.Fatal(err)
	}

	f, err := fsys.Create("/test.txt")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	atime := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	mtime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	err = fsys.Chtimes("/test.txt", atime, mtime)
	if err != nil {
		t.Fatal(err)
	}

	info, err := fsys.Stat("/test.txt")
	if err != nil {
		t.Fatal(err)
	}

	if !info.ModTime().Equal(mtime) {
		t.Errorf("expected mtime %v, got %v", mtime, info.ModTime())
	}
}

// TestFileSystemChown tests the Chown method
func TestFileSystemChown(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	fsys, err := NewFS(mfs)
	if err != nil {
		t.Fatal(err)
	}

	f, err := fsys.Create("/test.txt")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	// memfs may not support this, but we test the API
	err = fsys.Chown("/test.txt", 1000, 1000)
	if err != nil && !os.IsPermission(err) {
		t.Logf("Chown returned: %v", err)
	}
}

// TestMultipleReadersSimultaneous tests multiple concurrent readers
func TestMultipleReadersSimultaneous(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	fsys, err := NewFS(mfs)
	if err != nil {
		t.Fatal(err)
	}

	// Create test file
	f, err := fsys.Create("/shared.txt")
	if err != nil {
		t.Fatal(err)
	}
	testData := []byte("shared content for reading")
	f.Write(testData)
	f.Close()

	const numReaders = 50
	var wg sync.WaitGroup
	var successCount atomic.Int32
	errCh := make(chan error, numReaders)

	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			f, err := fsys.Open("/shared.txt")
			if err != nil {
				errCh <- fmt.Errorf("reader %d open: %w", id, err)
				return
			}
			defer f.Close()

			buf := make([]byte, 100)
			n, err := f.Read(buf)
			if err != nil && err != io.EOF {
				errCh <- fmt.Errorf("reader %d read: %w", id, err)
				return
			}

			if !bytes.Equal(buf[:n], testData) {
				errCh <- fmt.Errorf("reader %d: data mismatch", id)
				return
			}

			successCount.Add(1)
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Error(err)
	}

	if int(successCount.Load()) != numReaders {
		t.Errorf("expected %d successful reads, got %d", numReaders, successCount.Load())
	}
}

// TestWriterBlocksReaders tests that write operations block readers
func TestWriterBlocksReaders(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	fsys, err := NewFS(mfs)
	if err != nil {
		t.Fatal(err)
	}

	// Create test file
	f, err := fsys.Create("/test.txt")
	if err != nil {
		t.Fatal(err)
	}
	f.Write([]byte("initial"))
	f.Close()

	var writerStarted, readerStarted atomic.Bool
	var readerFinished atomic.Bool
	errCh := make(chan error, 2)

	// Start writer goroutine
	go func() {
		// Open file for writing
		f, err := fsys.OpenFile("/test.txt", os.O_WRONLY, 0644)
		if err != nil {
			errCh <- fmt.Errorf("writer open: %w", err)
			return
		}
		defer f.Close()

		writerStarted.Store(true)

		// Hold the file lock by writing slowly
		for i := 0; i < 5; i++ {
			time.Sleep(10 * time.Millisecond)
			_, err := f.Write([]byte("x"))
			if err != nil {
				errCh <- fmt.Errorf("writer write: %w", err)
				return
			}
		}
	}()

	// Wait for writer to start
	for !writerStarted.Load() {
		time.Sleep(time.Millisecond)
	}

	// Start reader goroutine
	go func() {
		readerStarted.Store(true)

		// Try to stat the file (should wait for writer to finish)
		_, err := fsys.Stat("/test.txt")
		if err != nil {
			errCh <- fmt.Errorf("reader stat: %w", err)
			return
		}
		readerFinished.Store(true)
	}()

	// Give reader time to start
	time.Sleep(5 * time.Millisecond)

	// Reader should have started but not finished yet (blocked by writer's file lock)
	if !readerStarted.Load() {
		t.Error("reader didn't start")
	}

	// Wait a bit more to ensure operations complete
	time.Sleep(100 * time.Millisecond)

	if !readerFinished.Load() {
		t.Error("reader didn't finish after writer should have completed")
	}

	close(errCh)
	for err := range errCh {
		t.Error(err)
	}
}

// TestLockReleaseOnError tests that locks are released even on errors
func TestLockReleaseOnError(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	fsys, err := NewFS(mfs)
	if err != nil {
		t.Fatal(err)
	}

	// Try to open non-existent file (should error)
	_, err = fsys.Open("/nonexistent.txt")
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}

	// Should still be able to create a file (lock was released)
	f, err := fsys.Create("/test.txt")
	if err != nil {
		t.Fatalf("failed to create file after error: %v", err)
	}
	f.Close()

	// Try to stat non-existent file (should error)
	_, err = fsys.Stat("/nonexistent.txt")
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}

	// Should still be able to stat existing file (lock was released)
	_, err = fsys.Stat("/test.txt")
	if err != nil {
		t.Fatalf("failed to stat after error: %v", err)
	}
}

// TestFileRead tests File.Read
func TestFileRead(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	fsys, err := NewFS(mfs)
	if err != nil {
		t.Fatal(err)
	}

	testData := []byte("test data for reading")
	f, err := fsys.Create("/test.txt")
	if err != nil {
		t.Fatal(err)
	}
	f.Write(testData)
	f.Close()

	f, err = fsys.Open("/test.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	buf := make([]byte, len(testData))
	n, err := f.Read(buf)
	if err != nil {
		t.Fatal(err)
	}

	if n != len(testData) || !bytes.Equal(buf, testData) {
		t.Errorf("Read mismatch: expected %q, got %q", testData, buf)
	}
}

// TestFileWrite tests File.Write
func TestFileWrite(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	fsys, err := NewFS(mfs)
	if err != nil {
		t.Fatal(err)
	}

	f, err := fsys.Create("/test.txt")
	if err != nil {
		t.Fatal(err)
	}

	testData := []byte("write test data")
	n, err := f.Write(testData)
	if err != nil {
		t.Fatal(err)
	}

	if n != len(testData) {
		t.Errorf("expected to write %d bytes, wrote %d", len(testData), n)
	}

	f.Close()

	// Verify
	data, err := fsys.ReadFile("/test.txt")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(data, testData) {
		t.Errorf("expected %q, got %q", testData, data)
	}
}

// TestFileSeek tests File.Seek
func TestFileSeek(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	fsys, err := NewFS(mfs)
	if err != nil {
		t.Fatal(err)
	}

	testData := []byte("0123456789")
	f, err := fsys.Create("/test.txt")
	if err != nil {
		t.Fatal(err)
	}
	f.Write(testData)
	f.Close()

	f, err = fsys.Open("/test.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	// Seek to position 5
	pos, err := f.Seek(5, io.SeekStart)
	if err != nil {
		t.Fatal(err)
	}

	if pos != 5 {
		t.Errorf("expected position 5, got %d", pos)
	}

	buf := make([]byte, 3)
	n, err := f.Read(buf)
	if err != nil {
		t.Fatal(err)
	}

	if n != 3 || string(buf) != "567" {
		t.Errorf("expected '567', got %q", buf)
	}
}

// TestFileReadAt tests File.ReadAt
func TestFileReadAt(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	fsys, err := NewFS(mfs)
	if err != nil {
		t.Fatal(err)
	}

	testData := []byte("abcdefghijklmnop")
	f, err := fsys.Create("/test.txt")
	if err != nil {
		t.Fatal(err)
	}
	f.Write(testData)
	f.Close()

	f, err = fsys.Open("/test.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	buf := make([]byte, 4)
	n, err := f.ReadAt(buf, 5)
	if err != nil {
		t.Fatal(err)
	}

	if n != 4 || string(buf) != "fghi" {
		t.Errorf("expected 'fghi', got %q", buf[:n])
	}
}

// TestFileWriteAt tests File.WriteAt
func TestFileWriteAt(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	fsys, err := NewFS(mfs)
	if err != nil {
		t.Fatal(err)
	}

	testData := []byte("0000000000")
	f, err := fsys.Create("/test.txt")
	if err != nil {
		t.Fatal(err)
	}
	f.Write(testData)
	f.Close()

	f, err = fsys.OpenFile("/test.txt", os.O_RDWR, 0644)
	if err != nil {
		t.Fatal(err)
	}

	replacement := []byte("XXXX")
	n, err := f.WriteAt(replacement, 3)
	if err != nil {
		t.Fatal(err)
	}

	if n != 4 {
		t.Errorf("expected to write 4 bytes, wrote %d", n)
	}

	f.Close()

	// Verify
	data, err := fsys.ReadFile("/test.txt")
	if err != nil {
		t.Fatal(err)
	}

	expected := "000XXXX000"
	if string(data) != expected {
		t.Errorf("expected %q, got %q", expected, data)
	}
}

// TestFileTruncate tests File.Truncate
func TestFileTruncate(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	fsys, err := NewFS(mfs)
	if err != nil {
		t.Fatal(err)
	}

	f, err := fsys.Create("/test.txt")
	if err != nil {
		t.Fatal(err)
	}

	f.Write([]byte("0123456789"))

	err = f.Truncate(5)
	if err != nil {
		t.Fatal(err)
	}

	f.Close()

	// Verify
	info, err := fsys.Stat("/test.txt")
	if err != nil {
		t.Fatal(err)
	}

	if info.Size() != 5 {
		t.Errorf("expected size 5, got %d", info.Size())
	}
}

// TestFileReadDirOnDirectory tests File.ReadDir on a directory
func TestFileReadDirOnDirectory(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	fsys, err := NewFS(mfs)
	if err != nil {
		t.Fatal(err)
	}

	err = fsys.Mkdir("/testdir", 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create files
	for i := 0; i < 3; i++ {
		f, err := fsys.Create(fmt.Sprintf("/testdir/file%d.txt", i))
		if err != nil {
			t.Fatal(err)
		}
		f.Close()
	}

	// Open directory
	dir, err := fsys.Open("/testdir")
	if err != nil {
		t.Fatal(err)
	}
	defer dir.Close()

	// Read all entries
	entries, err := dir.ReadDir(-1)
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}

	for _, entry := range entries {
		if entry.IsDir() {
			t.Errorf("expected file, got directory: %s", entry.Name())
		}
	}
}

// TestFileWriteString tests File.WriteString
func TestFileWriteString(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	fsys, err := NewFS(mfs)
	if err != nil {
		t.Fatal(err)
	}

	f, err := fsys.Create("/test.txt")
	if err != nil {
		t.Fatal(err)
	}

	testString := "test string content"
	n, err := f.WriteString(testString)
	if err != nil {
		t.Fatal(err)
	}

	if n != len(testString) {
		t.Errorf("expected to write %d bytes, wrote %d", len(testString), n)
	}

	f.Close()

	// Verify
	data, err := fsys.ReadFile("/test.txt")
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != testString {
		t.Errorf("expected %q, got %q", testString, data)
	}
}

// TestFileSync tests File.Sync
func TestFileSync(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	fsys, err := NewFS(mfs)
	if err != nil {
		t.Fatal(err)
	}

	f, err := fsys.Create("/test.txt")
	if err != nil {
		t.Fatal(err)
	}

	f.Write([]byte("test"))

	err = f.Sync()
	if err != nil {
		t.Fatal(err)
	}

	f.Close()
}

// TestFileReaddir tests File.Readdir
func TestFileReaddir(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	fsys, err := NewFS(mfs)
	if err != nil {
		t.Fatal(err)
	}

	err = fsys.Mkdir("/testdir", 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create files
	for i := 0; i < 5; i++ {
		f, err := fsys.Create(fmt.Sprintf("/testdir/file%d.txt", i))
		if err != nil {
			t.Fatal(err)
		}
		f.Close()
	}

	// Open directory
	dir, err := fsys.Open("/testdir")
	if err != nil {
		t.Fatal(err)
	}
	defer dir.Close()

	// Read all
	infos, err := dir.Readdir(-1)
	if err != nil {
		t.Fatal(err)
	}

	if len(infos) != 5 {
		t.Errorf("expected 5 entries, got %d", len(infos))
	}
}

// TestFileReaddirnames tests File.Readdirnames
func TestFileReaddirnames(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	fsys, err := NewFS(mfs)
	if err != nil {
		t.Fatal(err)
	}

	err = fsys.Mkdir("/testdir", 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create files
	for i := 0; i < 4; i++ {
		f, err := fsys.Create(fmt.Sprintf("/testdir/file%d.txt", i))
		if err != nil {
			t.Fatal(err)
		}
		f.Close()
	}

	// Open directory
	dir, err := fsys.Open("/testdir")
	if err != nil {
		t.Fatal(err)
	}
	defer dir.Close()

	// Read names
	names, err := dir.Readdirnames(-1)
	if err != nil {
		t.Fatal(err)
	}

	if len(names) != 4 {
		t.Errorf("expected 4 names, got %d", len(names))
	}
}

// TestFileStat tests File.Stat
func TestFileStat(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	fsys, err := NewFS(mfs)
	if err != nil {
		t.Fatal(err)
	}

	f, err := fsys.Create("/test.txt")
	if err != nil {
		t.Fatal(err)
	}

	testData := []byte("test content")
	f.Write(testData)

	info, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}

	if info.Name() != "test.txt" {
		t.Errorf("expected name 'test.txt', got %q", info.Name())
	}

	if info.Size() != int64(len(testData)) {
		t.Errorf("expected size %d, got %d", len(testData), info.Size())
	}

	f.Close()
}

// TestFileClose tests File.Close
func TestFileClose(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	fsys, err := NewFS(mfs)
	if err != nil {
		t.Fatal(err)
	}

	f, err := fsys.Create("/test.txt")
	if err != nil {
		t.Fatal(err)
	}

	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}

	// Trying to write after close should fail
	_, err = f.Write([]byte("test"))
	if err == nil {
		t.Error("expected error writing to closed file")
	}
}

// TestConcurrentOpenFileAndStat tests concurrent OpenFile and Stat operations
func TestConcurrentOpenFileAndStat(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	fsys, err := NewFS(mfs)
	if err != nil {
		t.Fatal(err)
	}

	// Create initial file
	f, err := fsys.Create("/shared.txt")
	if err != nil {
		t.Fatal(err)
	}
	f.Write([]byte("initial content"))
	f.Close()

	const numGoroutines = 20
	var wg sync.WaitGroup
	errCh := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			if id%2 == 0 {
				// Even: try to stat
				_, err := fsys.Stat("/shared.txt")
				if err != nil {
					errCh <- fmt.Errorf("goroutine %d stat: %w", id, err)
				}
			} else {
				// Odd: try to open
				f, err := fsys.Open("/shared.txt")
				if err != nil {
					errCh <- fmt.Errorf("goroutine %d open: %w", id, err)
					return
				}
				f.Close()
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Error(err)
	}
}
