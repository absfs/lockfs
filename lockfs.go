package lockfs

import (
	"io/fs"
	"os"
	"sync"
	"time"

	"github.com/absfs/absfs"
)

// Filer wraps an absfs.Filer with a RWMutex for thread-safe access.
// Read operations (Stat) use RLock for concurrent access, write operations use Lock.
// Files returned from OpenFile use hierarchical locking to coordinate with the Filer.
type Filer struct {
	fs absfs.Filer
	m  sync.RWMutex
}

// locker interface implementation for hierarchical locking
func (f *Filer) rlock()   { f.m.RLock() }
func (f *Filer) runlock() { f.m.RUnlock() }
func (f *Filer) lock()    { f.m.Lock() }
func (f *Filer) unlock()  { f.m.Unlock() }

// NewFiler creates a new thread-safe Filer wrapper.
func NewFiler(filer absfs.Filer) (*Filer, error) {
	return &Filer{fs: filer}, nil
}

// OpenFile opens a file using the given flags and the given mode.
// The returned File is wrapped for thread-safe access with hierarchical locking.
func (f *Filer) OpenFile(name string, flag int, perm os.FileMode) (absfs.File, error) {
	f.m.Lock()
	defer f.m.Unlock()
	file, err := f.fs.OpenFile(name, flag, perm)
	return wrapFile(f, file, err)
}

// Mkdir creates a directory in the filesystem, return an error if any
// happens.
func (f *Filer) Mkdir(name string, perm os.FileMode) error {
	f.m.Lock()
	defer f.m.Unlock()
	return f.fs.Mkdir(name, perm)
}

// Remove removes a file identified by name, returning an error, if any
// happens.
func (f *Filer) Remove(name string) error {
	f.m.Lock()
	defer f.m.Unlock()
	return f.fs.Remove(name)
}

// Rename renames (moves) oldpath to newpath.
func (f *Filer) Rename(oldpath, newpath string) error {
	f.m.Lock()
	defer f.m.Unlock()
	return f.fs.Rename(oldpath, newpath)
}

// Stat returns the FileInfo structure describing file. If there is an error,
// it will be of type *PathError.
func (f *Filer) Stat(name string) (os.FileInfo, error) {
	f.m.RLock()
	defer f.m.RUnlock()
	return f.fs.Stat(name)
}

// Chmod changes the mode of the named file to mode.
func (f *Filer) Chmod(name string, mode os.FileMode) error {
	f.m.Lock()
	defer f.m.Unlock()
	return f.fs.Chmod(name, mode)
}

// Chtimes changes the access and modification times of the named file.
func (f *Filer) Chtimes(name string, atime time.Time, mtime time.Time) error {
	f.m.Lock()
	defer f.m.Unlock()
	return f.fs.Chtimes(name, atime, mtime)
}

// Chown changes the owner and group ids of the named file.
func (f *Filer) Chown(name string, uid, gid int) error {
	f.m.Lock()
	defer f.m.Unlock()
	return f.fs.Chown(name, uid, gid)
}

// ReadDir reads the named directory and returns all its directory entries.
func (f *Filer) ReadDir(name string) ([]fs.DirEntry, error) {
	f.m.RLock()
	defer f.m.RUnlock()
	return f.fs.ReadDir(name)
}

// ReadFile reads the named file and returns its contents.
func (f *Filer) ReadFile(name string) ([]byte, error) {
	f.m.RLock()
	defer f.m.RUnlock()
	return f.fs.ReadFile(name)
}

// Sub returns a filesystem corresponding to the subtree rooted at dir.
func (f *Filer) Sub(dir string) (fs.FS, error) {
	f.m.RLock()
	defer f.m.RUnlock()
	return absfs.FilerToFS(f.fs, dir)
}

// FileSystem wraps an absfs.FileSystem with a RWMutex for thread-safe access.
// Read operations use RLock for concurrent access, write operations use Lock.
// Files returned from Open/Create/OpenFile use hierarchical locking to coordinate
// with the FileSystem, preventing races between file operations and filesystem mutations.
type FileSystem struct {
	m  sync.RWMutex
	fs absfs.FileSystem
}

// locker interface implementation for hierarchical locking
func (f *FileSystem) rlock()   { f.m.RLock() }
func (f *FileSystem) runlock() { f.m.RUnlock() }
func (f *FileSystem) lock()    { f.m.Lock() }
func (f *FileSystem) unlock()  { f.m.Unlock() }

// NewFS creates a new thread-safe FileSystem wrapper.
func NewFS(fs absfs.FileSystem) (*FileSystem, error) {
	return &FileSystem{fs: fs}, nil
}

// OpenFile opens a file using the given flags and the given mode.
// The returned File is wrapped for thread-safe access with hierarchical locking.
func (f *FileSystem) OpenFile(name string, flag int, perm os.FileMode) (absfs.File, error) {
	f.m.Lock()
	defer f.m.Unlock()
	file, err := f.fs.OpenFile(name, flag, perm)
	return wrapFile(f, file, err)
}

// Mkdir creates a directory in the filesystem, return an error if any
// happens.
func (f *FileSystem) Mkdir(name string, perm os.FileMode) error {
	f.m.Lock()
	defer f.m.Unlock()
	return f.fs.Mkdir(name, perm)
}

// Remove removes a file identified by name, returning an error, if any
// happens.
func (f *FileSystem) Remove(name string) error {
	f.m.Lock()
	defer f.m.Unlock()
	return f.fs.Remove(name)
}

// Rename renames (moves) oldpath to newpath.
func (f *FileSystem) Rename(oldpath, newpath string) error {
	f.m.Lock()
	defer f.m.Unlock()
	return f.fs.Rename(oldpath, newpath)
}

// Stat returns the FileInfo structure describing file. If there is an error,
// it will be of type *PathError.
func (f *FileSystem) Stat(name string) (os.FileInfo, error) {
	f.m.RLock()
	defer f.m.RUnlock()
	return f.fs.Stat(name)
}

// Chmod changes the mode of the named file to mode.
func (f *FileSystem) Chmod(name string, mode os.FileMode) error {
	f.m.Lock()
	defer f.m.Unlock()
	return f.fs.Chmod(name, mode)
}

// Chtimes changes the access and modification times of the named file.
func (f *FileSystem) Chtimes(name string, atime time.Time, mtime time.Time) error {
	f.m.Lock()
	defer f.m.Unlock()
	return f.fs.Chtimes(name, atime, mtime)
}

// Chown changes the owner and group ids of the named file.
func (f *FileSystem) Chown(name string, uid, gid int) error {
	f.m.Lock()
	defer f.m.Unlock()
	return f.fs.Chown(name, uid, gid)
}

// Chdir changes the current working directory.
func (f *FileSystem) Chdir(dir string) error {
	f.m.Lock()
	defer f.m.Unlock()
	return f.fs.Chdir(dir)
}

// Getwd returns the current working directory.
func (f *FileSystem) Getwd() (dir string, err error) {
	f.m.RLock()
	defer f.m.RUnlock()
	return f.fs.Getwd()
}

// TempDir returns the default directory for temporary files.
func (f *FileSystem) TempDir() string {
	return f.fs.TempDir()
}

// Open opens the named file for reading.
// The returned File is wrapped for thread-safe access with hierarchical locking.
func (f *FileSystem) Open(name string) (absfs.File, error) {
	f.m.RLock()
	defer f.m.RUnlock()
	file, err := f.fs.Open(name)
	return wrapFile(f, file, err)
}

// Create creates the named file, truncating it if it already exists.
// The returned File is wrapped for thread-safe access with hierarchical locking.
func (f *FileSystem) Create(name string) (absfs.File, error) {
	f.m.Lock()
	defer f.m.Unlock()
	file, err := f.fs.Create(name)
	return wrapFile(f, file, err)
}

// MkdirAll creates a directory named path, along with any necessary parents.
func (f *FileSystem) MkdirAll(name string, perm os.FileMode) error {
	f.m.Lock()
	defer f.m.Unlock()
	return f.fs.MkdirAll(name, perm)
}

// RemoveAll removes path and any children it contains.
func (f *FileSystem) RemoveAll(path string) (err error) {
	f.m.Lock()
	defer f.m.Unlock()
	return f.fs.RemoveAll(path)
}

// Truncate changes the size of the named file.
func (f *FileSystem) Truncate(name string, size int64) error {
	f.m.Lock()
	defer f.m.Unlock()
	return f.fs.Truncate(name, size)
}

// ReadDir reads the named directory and returns all its directory entries.
func (f *FileSystem) ReadDir(name string) ([]fs.DirEntry, error) {
	f.m.RLock()
	defer f.m.RUnlock()
	return f.fs.ReadDir(name)
}

// ReadFile reads the named file and returns its contents.
func (f *FileSystem) ReadFile(name string) ([]byte, error) {
	f.m.RLock()
	defer f.m.RUnlock()
	return f.fs.ReadFile(name)
}

// Sub returns a filesystem corresponding to the subtree rooted at dir.
func (f *FileSystem) Sub(dir string) (fs.FS, error) {
	f.m.RLock()
	defer f.m.RUnlock()
	return absfs.FilerToFS(f.fs, dir)
}

// SymlinkFileSystem wraps an absfs.SymlinkFileSystem with a RWMutex for thread-safe access.
// Read operations use RLock for concurrent access, write operations use Lock.
// Files returned from Open/Create/OpenFile use hierarchical locking to coordinate
// with the SymlinkFileSystem, preventing races between file operations and filesystem mutations.
type SymlinkFileSystem struct {
	m   sync.RWMutex
	sfs absfs.SymlinkFileSystem
}

// locker interface implementation for hierarchical locking
func (f *SymlinkFileSystem) rlock()   { f.m.RLock() }
func (f *SymlinkFileSystem) runlock() { f.m.RUnlock() }
func (f *SymlinkFileSystem) lock()    { f.m.Lock() }
func (f *SymlinkFileSystem) unlock()  { f.m.Unlock() }

// NewSymlinkFS creates a new thread-safe SymlinkFileSystem wrapper.
func NewSymlinkFS(fs absfs.SymlinkFileSystem) (*SymlinkFileSystem, error) {
	return &SymlinkFileSystem{sfs: fs}, nil
}

// OpenFile opens a file using the given flags and the given mode.
// The returned File is wrapped for thread-safe access with hierarchical locking.
func (f *SymlinkFileSystem) OpenFile(name string, flag int, perm os.FileMode) (absfs.File, error) {
	f.m.Lock()
	defer f.m.Unlock()
	file, err := f.sfs.OpenFile(name, flag, perm)
	return wrapFile(f, file, err)
}

// Mkdir creates a directory in the filesystem, return an error if any
// happens.
func (f *SymlinkFileSystem) Mkdir(name string, perm os.FileMode) error {
	f.m.Lock()
	defer f.m.Unlock()
	return f.sfs.Mkdir(name, perm)
}

// Remove removes a file identified by name, returning an error, if any
// happens.
func (f *SymlinkFileSystem) Remove(name string) error {
	f.m.Lock()
	defer f.m.Unlock()
	return f.sfs.Remove(name)
}

// Rename renames (moves) oldpath to newpath.
func (f *SymlinkFileSystem) Rename(oldpath, newpath string) error {
	f.m.Lock()
	defer f.m.Unlock()
	return f.sfs.Rename(oldpath, newpath)
}

// Stat returns the FileInfo structure describing file. If there is an error,
// it will be of type *PathError.
func (f *SymlinkFileSystem) Stat(name string) (os.FileInfo, error) {
	f.m.RLock()
	defer f.m.RUnlock()
	return f.sfs.Stat(name)
}

// Chmod changes the mode of the named file to mode.
func (f *SymlinkFileSystem) Chmod(name string, mode os.FileMode) error {
	f.m.Lock()
	defer f.m.Unlock()
	return f.sfs.Chmod(name, mode)
}

// Chtimes changes the access and modification times of the named file.
func (f *SymlinkFileSystem) Chtimes(name string, atime time.Time, mtime time.Time) error {
	f.m.Lock()
	defer f.m.Unlock()
	return f.sfs.Chtimes(name, atime, mtime)
}

// Chown changes the owner and group ids of the named file.
func (f *SymlinkFileSystem) Chown(name string, uid, gid int) error {
	f.m.Lock()
	defer f.m.Unlock()
	return f.sfs.Chown(name, uid, gid)
}

// Chdir changes the current working directory.
func (f *SymlinkFileSystem) Chdir(dir string) error {
	f.m.Lock()
	defer f.m.Unlock()
	return f.sfs.Chdir(dir)
}

// Getwd returns the current working directory.
func (f *SymlinkFileSystem) Getwd() (dir string, err error) {
	f.m.RLock()
	defer f.m.RUnlock()
	return f.sfs.Getwd()
}

// TempDir returns the default directory for temporary files.
func (f *SymlinkFileSystem) TempDir() string {
	return f.sfs.TempDir()
}

// Open opens the named file for reading.
// The returned File is wrapped for thread-safe access with hierarchical locking.
func (f *SymlinkFileSystem) Open(name string) (absfs.File, error) {
	f.m.RLock()
	defer f.m.RUnlock()
	file, err := f.sfs.Open(name)
	return wrapFile(f, file, err)
}

// Create creates the named file, truncating it if it already exists.
// The returned File is wrapped for thread-safe access with hierarchical locking.
func (f *SymlinkFileSystem) Create(name string) (absfs.File, error) {
	f.m.Lock()
	defer f.m.Unlock()
	file, err := f.sfs.Create(name)
	return wrapFile(f, file, err)
}

// MkdirAll creates a directory named path, along with any necessary parents.
func (f *SymlinkFileSystem) MkdirAll(name string, perm os.FileMode) error {
	f.m.Lock()
	defer f.m.Unlock()
	return f.sfs.MkdirAll(name, perm)
}

// RemoveAll removes path and any children it contains.
func (f *SymlinkFileSystem) RemoveAll(path string) (err error) {
	f.m.Lock()
	defer f.m.Unlock()
	return f.sfs.RemoveAll(path)
}

// Truncate changes the size of the named file.
func (f *SymlinkFileSystem) Truncate(name string, size int64) error {
	f.m.Lock()
	defer f.m.Unlock()
	return f.sfs.Truncate(name, size)
}

// ReadDir reads the named directory and returns all its directory entries.
func (f *SymlinkFileSystem) ReadDir(name string) ([]fs.DirEntry, error) {
	f.m.RLock()
	defer f.m.RUnlock()
	return f.sfs.ReadDir(name)
}

// ReadFile reads the named file and returns its contents.
func (f *SymlinkFileSystem) ReadFile(name string) ([]byte, error) {
	f.m.RLock()
	defer f.m.RUnlock()
	return f.sfs.ReadFile(name)
}

// Sub returns a filesystem corresponding to the subtree rooted at dir.
func (f *SymlinkFileSystem) Sub(dir string) (fs.FS, error) {
	f.m.RLock()
	defer f.m.RUnlock()
	return absfs.FilerToFS(f.sfs, dir)
}

// Lstat returns a FileInfo describing the named file. If the file is a
// symbolic link, the returned FileInfo describes the symbolic link. Lstat
// makes no attempt to follow the link. If there is an error, it will be of type *PathError.
func (f *SymlinkFileSystem) Lstat(name string) (os.FileInfo, error) {
	f.m.RLock()
	defer f.m.RUnlock()
	return f.sfs.Lstat(name)
}

// Lchown changes the numeric uid and gid of the named file. If the file is a
// symbolic link, it changes the uid and gid of the link itself. If there is
// an error, it will be of type *PathError.
//
// On Windows, it always returns the syscall.EWINDOWS error, wrapped in
// *PathError.
func (f *SymlinkFileSystem) Lchown(name string, uid, gid int) error {
	f.m.Lock()
	defer f.m.Unlock()
	return f.sfs.Lchown(name, uid, gid)
}

// Readlink returns the destination of the named symbolic link. If there is an
// error, it will be of type *PathError.
func (f *SymlinkFileSystem) Readlink(name string) (string, error) {
	f.m.RLock()
	defer f.m.RUnlock()
	return f.sfs.Readlink(name)
}

// Symlink creates newname as a symbolic link to oldname. If there is an
// error, it will be of type *LinkError.
func (f *SymlinkFileSystem) Symlink(oldname, newname string) error {
	f.m.Lock()
	defer f.m.Unlock()
	return f.sfs.Symlink(oldname, newname)
}
