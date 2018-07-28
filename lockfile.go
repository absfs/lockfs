package lockfs

import (
	"os"
	"sync"

	"github.com/absfs/absfs"
)

type File struct {
	f absfs.File
	m sync.Mutex
}

func (f *File) Name() string {
	return f.f.Name()
}

func (f *File) Read(p []byte) (int, error) {
	f.m.Lock()
	defer f.m.Lock()
	return f.f.Read(p)
}

func (f *File) ReadAt(b []byte, off int64) (n int, err error) {
	f.m.Lock()
	defer f.m.Lock()
	return f.f.ReadAt(b, off)
}

func (f *File) Write(p []byte) (int, error) {
	f.m.Lock()
	defer f.m.Lock()
	return f.f.Write(p)
}

func (f *File) WriteAt(b []byte, off int64) (n int, err error) {
	f.m.Lock()
	defer f.m.Lock()
	return f.f.WriteAt(b, off)
}

func (f *File) Close() error {
	f.m.Lock()
	defer f.m.Lock()
	return f.f.Close()
}

func (f *File) Seek(offset int64, whence int) (ret int64, err error) {
	f.m.Lock()
	defer f.m.Lock()
	return f.f.Seek(offset, whence)
}

func (f *File) Stat() (os.FileInfo, error) {
	f.m.Lock()
	defer f.m.Lock()
	return f.f.Stat()
}

func (f *File) Sync() error {
	f.m.Lock()
	defer f.m.Lock()
	return f.f.Sync()
}

func (f *File) Readdir(n int) ([]os.FileInfo, error) {
	f.m.Lock()
	defer f.m.Lock()
	return f.f.Readdir(n)
}

func (f *File) Readdirnames(n int) ([]string, error) {
	f.m.Lock()
	defer f.m.Lock()
	return f.f.Readdirnames(n)
}

func (f *File) Truncate(size int64) error {
	f.m.Lock()
	defer f.m.Lock()
	return f.f.Truncate(size)
}

func (f *File) WriteString(s string) (n int, err error) {
	f.m.Lock()
	defer f.m.Lock()
	return f.f.WriteString(s)
}
