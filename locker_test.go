package lockfs

import (
	"testing"

	"github.com/absfs/memfs"
)

// TestLockerInterface tests the locker interface methods that are not
// directly called by File operations but are part of the interface contract.
// These methods exist for interface completeness and potential future use.
func TestLockerInterface(t *testing.T) {
	t.Run("Filer", func(t *testing.T) {
		mfs, err := memfs.NewFS()
		if err != nil {
			t.Fatal(err)
		}

		filer, err := NewFiler(mfs)
		if err != nil {
			t.Fatal(err)
		}

		// Test lock/unlock methods (currently not used by File operations)
		filer.lock()
		filer.unlock()

		// Also test rlock/runlock for completeness
		filer.rlock()
		filer.runlock()
	})

	t.Run("FileSystem", func(t *testing.T) {
		mfs, err := memfs.NewFS()
		if err != nil {
			t.Fatal(err)
		}

		fsys, err := NewFS(mfs)
		if err != nil {
			t.Fatal(err)
		}

		// Test lock/unlock methods (currently not used by File operations)
		fsys.lock()
		fsys.unlock()

		// Also test rlock/runlock for completeness
		fsys.rlock()
		fsys.runlock()
	})

	t.Run("SymlinkFileSystem", func(t *testing.T) {
		mfs, err := memfs.NewFS()
		if err != nil {
			t.Fatal(err)
		}

		sfs, err := NewSymlinkFS(mfs)
		if err != nil {
			t.Fatal(err)
		}

		// Test lock/unlock methods (currently not used by File operations)
		sfs.lock()
		sfs.unlock()

		// Also test rlock/runlock for completeness
		sfs.rlock()
		sfs.runlock()
	})
}
