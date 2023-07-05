// Helper pachage for types and functions
package repo

import "os"

type FileInfo struct {
	F *os.File // file pointer
	O int64    // offset
}

func NewFileInfo(f *os.File, o int64) *FileInfo {
	return &FileInfo{
		F: f,
		O: o,
	}
}
func (f *FileInfo) AddOffset(o int64) {
	f.O += o
}
