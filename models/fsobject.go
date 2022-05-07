package models

import "strings"

const (
	S_IFMT   = 0o0170000
	S_IFSOCK = 0o0140000
	S_IFLNK  = 0o0120000
	S_IFREG  = 0o0100000
	S_IFBLK  = 0o0060000
	S_IFDIR  = 0o0040000
	S_IFCHR  = 0o0020000
	S_IFIFO  = 0o0010000
)

type FsObject struct {
	ID       uint64
	Path     string
	Hash     string
	Created  int64
	Modified int64
	Uid      uint32
	Gid      uint32
	Mode     uint32
	AgentID  uint64
}

func (f *FsObject) ParseMode() string {
	masks := map[uint32]byte{
		S_IFSOCK: 's',
		S_IFLNK:  'l',
		S_IFREG:  '-',
		S_IFBLK:  'b',
		S_IFDIR:  'd',
		S_IFCHR:  'c',
		S_IFIFO:  'p',
	}

	var sb strings.Builder
	sb.Grow(10)

	for mask, c := range masks {
		if f.Mode&S_IFMT == mask {
			sb.WriteByte(c)
			break
		}
	}

	const perms = "rwxrwxrwx"
	for i, c := range perms {
		switch 8 - i {
		case 6:
			if f.Mode&(1<<11) != 0 {
				sb.WriteByte('s')
				continue
			}
		case 3:
			if f.Mode&(1<<10) != 0 {
				if f.Mode&S_IFMT == S_IFDIR {
					sb.WriteByte('s')
				} else {
					sb.WriteByte('S')
				}
				continue
			}
		case 0:
			if f.Mode&(1<<9) != 0 {
				if f.Mode&S_IFMT == S_IFDIR {
					sb.WriteByte('t')
				} else {
					sb.WriteByte('T')
				}
				continue
			}
		}

		if f.Mode&(1<<uint32(8-i)) != 0 {
			sb.WriteByte(byte(c))
		} else {
			sb.WriteByte('-')
		}
	}

	return sb.String()
}
