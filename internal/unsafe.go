package internal

import "unsafe"

func String(raw []byte) string {
	return unsafe.String(unsafe.SliceData(raw), len(raw))
}
