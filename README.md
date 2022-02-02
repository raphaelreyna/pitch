# PitCH
Like tar but different.
PitCH is an archive file format that aims for high performance and and minimal bloat.

## PitCH vs tar
- PitCH header sections add at least 3 bytes to the archive and grow as `O(log(n))` where `n` is file name size or file content size.
This is because pitch only stores the file name and size whereas tar has a fixed 512 byte header that includes permissions, user, etc.

- PitCH has dynamically sized headers which means there is no limit to file name length or file content length; tars fixed header size limits both file name and file size.


### CLI util
The cli tool follows the tar command as closely as possible.
#### Examples
Archiving the directory `./mydir`
```sh
pitch -c -f mydir.pch ./mydir
```

Extracting an archive into `./mydir`
```sh
pitch -x -f mydir.pch -C ./mydir
```