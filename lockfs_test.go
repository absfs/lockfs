package lockfs

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/absfs/memfs"
)

// func TestLockfs(t *testing.T) {
// 	ofs := osfs.New()
// 	fs, err := NewFS(ofs)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// }

func TestMkdir(t *testing.T) {

	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	fs, err := NewFS(mfs)
	if err != nil {
		t.Fatal(err)
	}

	if fs.TempDir() != "/tmp" {
		t.Fatalf("wrong TempDir output: %q != %q", fs.TempDir(), "/tmp")
	}

	testdir := fs.TempDir()

	t.Logf("Test path: %q", testdir)
	err = fs.MkdirAll(testdir, 0777)
	if err != nil {
		t.Fatal(err)
	}

	var list []string
	path := "/"
outer:
	for _, name := range strings.Split(testdir, "/")[1:] {
		if name == "" {
			continue
		}
		f, err := fs.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		list, err = f.Readdirnames(-1)
		f.Close()
		if err != nil {
			t.Fatal(err)
		}
		for _, n := range list {
			if n == name {
				path = filepath.Join(path, name)
				continue outer
			}
		}
		t.Errorf("path error: %q + %q:  %s", path, name, list)
	}

}

func TestOpenWrite(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	fs, err := NewFS(mfs)
	if err != nil {
		t.Fatal(err)
	}

	f, err := fs.Create("/test_file.txt")
	if err != nil {
		t.Fatal(err)
	}

	data := []byte("The quick brown fox jumped over the lazy dog.\n")
	n, err := f.Write(data)
	f.Close()
	if n != len(data) {
		t.Errorf("write error: wrong byte count %d, expected %d", n, len(data))
	}
	if err != nil {
		t.Fatal(err)
	}

	f, err = fs.Open("/test_file.txt")
	if err != nil {
		t.Fatal(err)
	}
	buff := make([]byte, 512)
	n, err = f.Read(buff)
	f.Close()
	if n != len(data) {
		t.Errorf("write error: wrong byte count %d, expected %d", n, len(data))
	}
	if err != nil {
		t.Fatal(err)
	}
	buff = buff[:n]
	if bytes.Compare(data, buff) != 0 {
		t.Log(string(data))
		t.Log(string(buff))

		t.Fatal("bytes written do not compare to bytes read")
	}

}

func TestRename(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	fs, err := NewFS(mfs)
	if err != nil {
		t.Fatal(err)
	}

	// Create a file
	f, err := fs.Create("/original.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.Write([]byte("test content"))
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	// Rename it
	err = fs.Rename("/original.txt", "/renamed.txt")
	if err != nil {
		t.Fatal(err)
	}

	// Verify old file doesn't exist
	_, err = fs.Stat("/original.txt")
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}

	// Verify new file exists
	info, err := fs.Stat("/renamed.txt")
	if err != nil {
		t.Fatal(err)
	}
	if info.Name() != "renamed.txt" {
		t.Fatalf("unexpected name: %s", info.Name())
	}
}

func TestConcurrentFileSystemOperations(t *testing.T) {
	// Note: This test demonstrates that lockfs serializes filesystem operations.
	// The underlying memfs is not thread-safe, but lockfs wrapping ensures
	// that only one goroutine accesses memfs at a time for filesystem-level operations.
	//
	// However, once a File is returned, operations on that File may race with
	// other operations if the underlying filesystem stores shared state that
	// File operations access. This is a limitation of the decorator pattern -
	// true thread safety requires the underlying filesystem to be thread-safe,
	// or all File operations to also hold the filesystem lock.

	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	fs, err := NewFS(mfs)
	if err != nil {
		t.Fatal(err)
	}

	// Create base directory
	err = fs.MkdirAll("/concurrent", 0755)
	if err != nil {
		t.Fatal(err)
	}

	const numGoroutines = 10
	const numOps = 20

	var wg sync.WaitGroup
	errCh := make(chan error, numGoroutines*numOps)

	// Test concurrent Stat operations (read-only, should be safe)
	// First create all files sequentially
	for i := 0; i < numGoroutines; i++ {
		for j := 0; j < numOps; j++ {
			filename := fmt.Sprintf("/concurrent/file_%d_%d.txt", i, j)
			f, err := fs.Create(filename)
			if err != nil {
				t.Fatalf("create %s: %v", filename, err)
			}
			f.Write([]byte(fmt.Sprintf("content from goroutine %d, op %d", i, j)))
			f.Close()
		}
	}

	// Now test concurrent read operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				filename := fmt.Sprintf("/concurrent/file_%d_%d.txt", id, j)

				// Stat the file (read operation)
				_, err := fs.Stat(filename)
				if err != nil {
					errCh <- fmt.Errorf("stat %s: %w", filename, err)
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

func TestConcurrentFileOperations(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	fs, err := NewFS(mfs)
	if err != nil {
		t.Fatal(err)
	}

	// Create a shared file
	f, err := fs.Create("/shared.txt")
	if err != nil {
		t.Fatal(err)
	}

	data := []byte("initial content for concurrent access testing")
	_, err = f.Write(data)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	const numGoroutines = 10
	const numOps = 50

	var wg sync.WaitGroup
	errCh := make(chan error, numGoroutines*numOps)

	// Open the file once per goroutine and perform concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			f, err := fs.Open("/shared.txt")
			if err != nil {
				errCh <- fmt.Errorf("goroutine %d open: %w", id, err)
				return
			}
			defer f.Close()

			for j := 0; j < numOps; j++ {
				// Test Stat (read operation)
				_, err := f.Stat()
				if err != nil {
					errCh <- fmt.Errorf("goroutine %d stat: %w", id, err)
				}

				// Test Name (no lock needed)
				name := f.Name()
				if name != "/shared.txt" {
					errCh <- fmt.Errorf("goroutine %d unexpected name: %s", id, name)
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

func TestConcurrentReadAt(t *testing.T) {
	// Note: This test verifies that each File handle's operations are properly
	// serialized by its own lock. The underlying memfs is not thread-safe,
	// so sharing a single underlying File between goroutines would race.
	// Each goroutine opens its own file handle for proper isolation.

	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	fs, err := NewFS(mfs)
	if err != nil {
		t.Fatal(err)
	}

	// Create a file with known content
	f, err := fs.Create("/readat.txt")
	if err != nil {
		t.Fatal(err)
	}

	content := "0123456789ABCDEFGHIJ"
	_, err = f.Write([]byte(content))
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	const numGoroutines = 10
	var wg sync.WaitGroup
	errCh := make(chan error, numGoroutines)

	// Each goroutine opens its own file handle
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Each goroutine opens its own file
			f, err := fs.Open("/readat.txt")
			if err != nil {
				errCh <- fmt.Errorf("goroutine %d open: %w", id, err)
				return
			}
			defer f.Close()

			buf := make([]byte, 5)
			offset := int64((id % 4) * 5) // Offsets: 0, 5, 10, 15

			n, err := f.ReadAt(buf, offset)
			if err != nil {
				errCh <- fmt.Errorf("goroutine %d ReadAt: %w", id, err)
				return
			}

			expected := content[offset : offset+5]
			if string(buf[:n]) != expected {
				errCh <- fmt.Errorf("goroutine %d: expected %q, got %q", id, expected, string(buf[:n]))
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Error(err)
	}
}

func TestConcurrentMkdirAndRemove(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	fs, err := NewFS(mfs)
	if err != nil {
		t.Fatal(err)
	}

	const numGoroutines = 5
	const numOps = 20

	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numOps; j++ {
				dirName := fmt.Sprintf("/dir_%d_%d", id, j)

				// Create directory
				err := fs.Mkdir(dirName, 0755)
				if err != nil {
					t.Logf("mkdir %s: %v", dirName, err)
					continue
				}

				// Remove directory
				err = fs.Remove(dirName)
				if err != nil {
					t.Logf("remove %s: %v", dirName, err)
				}
			}
		}(i)
	}

	wg.Wait()
}

func TestHierarchicalLocking(t *testing.T) {
	// This test verifies that file operations hold the filesystem read lock,
	// preventing filesystem mutations from racing with file I/O.
	//
	// The hierarchical locking ensures:
	// 1. File read operations hold filesystem RLock (multiple readers OK)
	// 2. File write operations hold filesystem RLock (blocked by fs mutations)
	// 3. Filesystem mutations hold filesystem Lock (blocked by file ops)

	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	fs, err := NewFS(mfs)
	if err != nil {
		t.Fatal(err)
	}

	// Create a test file
	f, err := fs.Create("/testfile.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.Write([]byte("initial content"))
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	// Open the file for reading
	f, err = fs.Open("/testfile.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	const iterations = 100
	var wg sync.WaitGroup
	errCh := make(chan error, iterations*2)

	// Goroutine 1: Repeatedly read from the open file handle
	wg.Add(1)
	go func() {
		defer wg.Done()
		buf := make([]byte, 100)
		for i := 0; i < iterations; i++ {
			// Seek back to start and read
			_, err := f.Seek(0, 0)
			if err != nil {
				errCh <- fmt.Errorf("seek error: %w", err)
				return
			}
			_, err = f.Read(buf)
			if err != nil {
				errCh <- fmt.Errorf("read error: %w", err)
				return
			}
		}
	}()

	// Goroutine 2: Repeatedly stat the file (filesystem operation)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			_, err := fs.Stat("/testfile.txt")
			if err != nil {
				errCh <- fmt.Errorf("stat error: %w", err)
				return
			}
		}
	}()

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Error(err)
	}
}

func TestFilerInterface(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	// Get underlying filer from memfs
	filer, err := NewFiler(mfs)
	if err != nil {
		t.Fatal(err)
	}

	// Test basic operations
	err = filer.Mkdir("/test", 0755)
	if err != nil {
		t.Fatal(err)
	}

	f, err := filer.OpenFile("/test/file.txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	info, err := filer.Stat("/test/file.txt")
	if err != nil {
		t.Fatal(err)
	}
	if info.Name() != "file.txt" {
		t.Fatalf("unexpected name: %s", info.Name())
	}

	err = filer.Rename("/test/file.txt", "/test/renamed.txt")
	if err != nil {
		t.Fatal(err)
	}

	_, err = filer.Stat("/test/renamed.txt")
	if err != nil {
		t.Fatal(err)
	}

	err = filer.Remove("/test/renamed.txt")
	if err != nil {
		t.Fatal(err)
	}
}
