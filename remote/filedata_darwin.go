package remote

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

func getFileMetadata(path string, info os.FileInfo) map[string]interface{} {

	hardlink := false
	symlink := false
	var linkTarget string

	var inodedata interface{}
	if inode, err := getInodeinfo(info); err != nil {
		inodedata = map[string]interface{}{"error": err.Error()}
	} else {
		inodedata = inode
		if inode.NLink > 1 {
			hardlink = true
		}
	}
	if info.Mode()&os.ModeSymlink == os.ModeSymlink {
		symlink = true
		linkTarget, _ = os.Readlink(path)
	}
	m := map[string]interface{}{
		"path":        filepath.Clean(path),
		"is_dir":      info.IsDir(),
		"dir":         filepath.Dir(path),
		"name":        info.Name(),
		"mode":        fmt.Sprintf("%o", info.Mode()),
		"mode_d":      fmt.Sprintf("%v", uint32(info.Mode())),
		"perm":        info.Mode().String(),
		"inode":       inodedata,
		"size":        info.Size(),
		"is_hardlink": hardlink,
		"is_symlink":  symlink,
		"link":        linkTarget,
	}

	if stat := info.Sys().(*syscall.Stat_t); stat != nil {
		m["atime"] = time.Unix(stat.Atimespec.Sec, stat.Atimespec.Nsec).Format(time.RFC3339Nano)
		m["mtime"] = time.Unix(stat.Mtimespec.Sec, stat.Mtimespec.Nsec).Format(time.RFC3339Nano)
		m["uid"] = stat.Uid
		m["gid"] = stat.Gid
	}

	ext := filepath.Ext(info.Name())
	if len(ext) > 0 {
		m["ext"] = ext
	}

	return m
}

type inodeinfo struct {
	// NLink is the number of times this file is linked to by
	// hardlinks.
	NLink uint16
	// Ino is the inode number for the file.
	Ino uint64
}

func getInodeinfo(fi os.FileInfo) (*inodeinfo, error) {
	var statT *syscall.Stat_t
	var ok bool
	if statT, ok = fi.Sys().(*syscall.Stat_t); !ok {
		return nil, errors.New("unable to determine if file is a hardlink (expected syscall.Stat_t)")
	}
	return &inodeinfo{
		Ino:   statT.Ino,
		NLink: statT.Nlink,
	}, nil
}
