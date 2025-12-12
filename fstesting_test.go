package lockfs

import (
	"testing"

	"github.com/absfs/absfs"
	"github.com/absfs/fstesting"
	"github.com/absfs/memfs"
)

// TestLockFSWrapper verifies that lockfs correctly wraps a base filesystem
// with thread-safe locking without transforming data or metadata.
func TestLockFSWrapper(t *testing.T) {
	baseFS, err := memfs.NewFS()
	if err != nil {
		t.Fatalf("failed to create base filesystem: %v", err)
	}

	suite := &fstesting.WrapperSuite{
		Factory: func(base absfs.FileSystem) (absfs.FileSystem, error) {
			return NewFS(base)
		},
		BaseFS:         baseFS,
		Name:           "lockfs",
		TransformsData: false,
		TransformsMeta: false,
		ReadOnly:       false,
	}

	suite.Run(t)
}

// TestSymlinkFSSuite runs the fstesting suite against the SymlinkFileSystem wrapper.
// This tests that the SymlinkFileSystem wrapper correctly delegates all operations
// including symlink-specific operations to the underlying filesystem.
func TestSymlinkFSSuite(t *testing.T) {
	baseFS, err := memfs.NewFS()
	if err != nil {
		t.Fatalf("failed to create base filesystem: %v", err)
	}

	sfs, err := NewSymlinkFS(baseFS)
	if err != nil {
		t.Fatalf("failed to create SymlinkFileSystem: %v", err)
	}

	suite := &fstesting.Suite{
		FS: sfs,
		Features: fstesting.Features{
			Symlinks:      true,
			HardLinks:     false, // memfs doesn't support hard links
			Permissions:   true,
			Timestamps:    true,
			CaseSensitive: true,
			AtomicRename:  true,
			SparseFiles:   false,
			LargeFiles:    true,
		},
	}

	suite.Run(t)
}
