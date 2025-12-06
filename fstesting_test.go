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
