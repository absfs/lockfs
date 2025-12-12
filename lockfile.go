package lockfs

import (
	"io/fs"
	"os"
	"sync"

	"github.com/absfs/absfs"
)

// locker is an interface for types that provide filesystem-level locking.
// File uses this to coordinate with its parent filesystem wrapper.
type locker interface {
	rlock()
	runlock()
	lock()
	unlock()
}

// File wraps an absfs.File with hierarchical locking for thread-safe access.
//
// File operations acquire both:
// 1. A read lock on the parent filesystem (prevents filesystem mutations during I/O)
// 2. An appropriate lock on the file itself (serializes operations on this handle)
//
// This ensures that operations like fs.Create("/file") cannot race with
// f.Read() on an existing handle to the same file.
type File struct {
	f      absfs.File
	m      sync.RWMutex
	parent locker
}

// wrapFile wraps an absfs.File in a thread-safe File wrapper with hierarchical locking.
// The parent parameter provides the filesystem-level lock for coordination.
func wrapFile(parent locker, f absfs.File, err error) (absfs.File, error) {
	if err != nil {
		return nil, err
	}
	return &File{f: f, parent: parent}, nil
}

// Name returns the name of the file. This is safe without locking
// since the name is immutable after file creation.
func (f *File) Name() string {
	return f.f.Name()
}

// Read reads up to len(p) bytes into p.
// Uses exclusive file lock (modifies position) with filesystem read lock.
func (f *File) Read(p []byte) (int, error) {
	f.parent.rlock()
	defer f.parent.runlock()
	f.m.Lock()
	defer f.m.Unlock()
	return f.f.Read(p)
}

// ReadAt reads len(b) bytes from the file starting at byte offset off.
// Uses read locks on both filesystem and file (position-independent).
func (f *File) ReadAt(b []byte, off int64) (n int, err error) {
	f.parent.rlock()
	defer f.parent.runlock()
	f.m.RLock()
	defer f.m.RUnlock()
	return f.f.ReadAt(b, off)
}

// Write writes len(p) bytes to the file.
// Uses exclusive locks on both filesystem and file.
func (f *File) Write(p []byte) (int, error) {
	f.parent.rlock()
	defer f.parent.runlock()
	f.m.Lock()
	defer f.m.Unlock()
	return f.f.Write(p)
}

// WriteAt writes len(b) bytes to the file starting at byte offset off.
// Uses filesystem read lock and exclusive file lock.
func (f *File) WriteAt(b []byte, off int64) (n int, err error) {
	f.parent.rlock()
	defer f.parent.runlock()
	f.m.Lock()
	defer f.m.Unlock()
	return f.f.WriteAt(b, off)
}

// Close closes the file.
// Uses exclusive file lock only (closing doesn't need filesystem lock).
func (f *File) Close() error {
	f.m.Lock()
	defer f.m.Unlock()
	return f.f.Close()
}

// Seek sets the offset for the next Read or Write.
// Uses exclusive file lock (modifies position) with filesystem read lock.
func (f *File) Seek(offset int64, whence int) (ret int64, err error) {
	f.parent.rlock()
	defer f.parent.runlock()
	f.m.Lock()
	defer f.m.Unlock()
	return f.f.Seek(offset, whence)
}

// Stat returns the FileInfo for the file.
// Uses read locks on both filesystem and file.
func (f *File) Stat() (os.FileInfo, error) {
	f.parent.rlock()
	defer f.parent.runlock()
	f.m.RLock()
	defer f.m.RUnlock()
	return f.f.Stat()
}

// Sync commits the file's contents to stable storage.
// Uses filesystem read lock and exclusive file lock.
func (f *File) Sync() error {
	f.parent.rlock()
	defer f.parent.runlock()
	f.m.Lock()
	defer f.m.Unlock()
	return f.f.Sync()
}

// Readdir reads the contents of the directory.
// Uses exclusive file lock (modifies directory cursor) with filesystem read lock.
func (f *File) Readdir(n int) ([]os.FileInfo, error) {
	f.parent.rlock()
	defer f.parent.runlock()
	f.m.Lock()
	defer f.m.Unlock()
	return f.f.Readdir(n)
}

// Readdirnames reads the names of directory entries.
// Uses exclusive file lock (modifies directory cursor) with filesystem read lock.
func (f *File) Readdirnames(n int) ([]string, error) {
	f.parent.rlock()
	defer f.parent.runlock()
	f.m.Lock()
	defer f.m.Unlock()
	return f.f.Readdirnames(n)
}

// Truncate changes the size of the file.
// Uses filesystem read lock and exclusive file lock.
func (f *File) Truncate(size int64) error {
	f.parent.rlock()
	defer f.parent.runlock()
	f.m.Lock()
	defer f.m.Unlock()
	return f.f.Truncate(size)
}

// WriteString writes a string to the file.
// Uses filesystem read lock and exclusive file lock.
func (f *File) WriteString(s string) (n int, err error) {
	f.parent.rlock()
	defer f.parent.runlock()
	f.m.Lock()
	defer f.m.Unlock()
	return f.f.WriteString(s)
}

// ReadDir reads the contents of the directory associated with file and
// returns a slice of up to n DirEntry values, as would be returned
// by ReadDir. If n <= 0, ReadDir returns all the DirEntry values from
// the directory in a single slice.
// Uses exclusive file lock (modifies directory cursor) with filesystem read lock.
func (f *File) ReadDir(n int) ([]fs.DirEntry, error) {
	f.parent.rlock()
	defer f.parent.runlock()
	f.m.Lock()
	defer f.m.Unlock()
	return f.f.ReadDir(n)
}
