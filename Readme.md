# lockfs - A thread safe wrapper for absfs.FileSystem
The lockfs FileSystem simply wraps all methods for absfs interfaces with a mutex Lock/Unlock making files and filesystems thread safe.

It's a good simple example of how functionality is added by composition of absfs.FileSystems.

## absfs
Check out the [`absfs`](https://github.com/absfs/absfs) repo for more information about the abstract FileSystem interface and features like FileSystem composition.

## LICENSE

This project is governed by the MIT License. See [LICENSE](https://github.com/absfs/lockfs/blob/master/LICENSE)



