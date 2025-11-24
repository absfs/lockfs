# lockfs - Thread-Safe Wrapper for absfs

[![Go Reference](https://pkg.go.dev/badge/github.com/absfs/lockfs.svg)](https://pkg.go.dev/github.com/absfs/lockfs)

The `lockfs` package provides thread-safe wrappers for `absfs` filesystem interfaces. It uses hierarchical `sync.RWMutex` locking to enable safe concurrent access from multiple goroutines while preventing races between file operations and filesystem mutations.

## Features

- **Full absfs interface compliance**: Implements `Filer`, `FileSystem`, and `SymlinkFileSystem` interfaces
- **Hierarchical locking**: File operations hold a read lock on the parent filesystem, preventing races with filesystem mutations
- **RWMutex-based**: Read operations allow concurrent access; write operations use exclusive locks
- **Wrapped file handles**: Files returned from Open/Create/OpenFile are automatically wrapped for thread-safe access
- **Zero dependencies** beyond absfs itself

## Installation

```bash
go get github.com/absfs/lockfs
```

## Usage

### Wrapping a FileSystem

```go
package main

import (
    "github.com/absfs/lockfs"
    "github.com/absfs/memfs"
)

func main() {
    // Create an underlying filesystem
    mfs, _ := memfs.NewFS()

    // Wrap it for thread-safe access
    fs, _ := lockfs.NewFS(mfs)

    // Now safe to use from multiple goroutines
    go func() {
        fs.Create("/file1.txt")
    }()
    go func() {
        fs.Create("/file2.txt")
    }()
}
```

### Wrapping a Filer

```go
filer, _ := lockfs.NewFiler(myFiler)
```

### Wrapping a SymlinkFileSystem

```go
sfs, _ := lockfs.NewSymlinkFS(mySymlinkFS)
```

## Hierarchical Locking

The key feature of `lockfs` is its hierarchical locking strategy. When you perform a file operation (like `Read` or `Write`), the operation acquires:

1. **Filesystem read lock** - Prevents filesystem mutations (Create, Remove, Rename, etc.) from executing concurrently
2. **File-level lock** - Serializes operations on the specific file handle

This design prevents races between scenarios like:
- Reading from an open file while another goroutine removes or truncates it
- Writing to a file while another goroutine renames it
- Multiple file handles to the same path operating concurrently

```
Goroutine A: f.Read()          Goroutine B: fs.Remove("/file")
    |                               |
    +-- fs.RLock()                  +-- fs.Lock() [BLOCKED]
    +-- file.Lock()                 |
    +-- ... read ...                |
    +-- file.Unlock()               |
    +-- fs.RUnlock()                |
                                    +-- [NOW PROCEEDS]
```

## Thread Safety Semantics

### Filesystem Operations

All filesystem operations are protected by a `RWMutex`:

| Operation | Lock Type | Notes |
|-----------|-----------|-------|
| Stat, Lstat | RLock | Concurrent reads allowed |
| Open | RLock | Concurrent opens allowed |
| Getwd | RLock | Concurrent reads allowed |
| Readlink | RLock | Concurrent reads allowed |
| Create, OpenFile | Lock | Exclusive access |
| Mkdir, MkdirAll | Lock | Exclusive access |
| Remove, RemoveAll | Lock | Exclusive access |
| Rename | Lock | Exclusive access |
| Chmod, Chown, Chtimes | Lock | Exclusive access |
| Chdir | Lock | Exclusive access |
| Truncate | Lock | Exclusive access |
| Symlink | Lock | Exclusive access |

### File Operations

Each `File` has its own `RWMutex` plus holds the parent filesystem's read lock during operations:

| Operation | FS Lock | File Lock | Notes |
|-----------|---------|-----------|-------|
| Name | None | None | Immutable after creation |
| Stat | RLock | RLock | Concurrent reads allowed |
| ReadAt | RLock | RLock | Position-independent read |
| Read | RLock | Lock | Modifies file position |
| Write, WriteAt, WriteString | RLock | Lock | Exclusive file access |
| Seek | RLock | Lock | Modifies file position |
| Truncate | RLock | Lock | Exclusive file access |
| Readdir, Readdirnames | RLock | Lock | Modifies directory cursor |
| Sync | RLock | Lock | Exclusive file access |
| Close | None | Lock | No filesystem lock needed |

## Lock Ordering

To prevent deadlocks, locks are always acquired in this order:
1. Filesystem lock (if needed)
2. File lock (if needed)

And released in reverse order via `defer`.

## Limitations

### Underlying Filesystem Thread Safety

The `lockfs` wrapper serializes access at the lockfs level, but cannot fix thread-safety issues within the underlying filesystem's implementation. If the underlying filesystem has internal races, those may still occur.

**Best practice**: Use `lockfs` to add thread safety to single-threaded filesystem implementations.

### Performance Considerations

- File operations hold the filesystem read lock, which blocks filesystem mutations
- Multiple concurrent reads are efficient (RLock allows multiple readers)
- Write-heavy workloads may see contention on the filesystem lock
- Consider using separate filesystem instances for isolated workloads

## absfs

Check out the [`absfs`](https://github.com/absfs/absfs) repo for more information about the abstract FileSystem interface and features like FileSystem composition.

## License

This project is governed by the MIT License. See [LICENSE](https://github.com/absfs/lockfs/blob/master/LICENSE)
