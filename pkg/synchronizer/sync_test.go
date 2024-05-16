package synchronizer

import (
	"context"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

var (
	fs = afero.NewOsFs()
)

func TestSyncDirs(t *testing.T) {
	fs.MkdirAll("/src/subdir", 0644)
	fs.MkdirAll("/dst", 0644)
	fs.Create("/src/subdir/file")
	fs.Create("/dst/file")
	afero.WriteFile(fs, "/src/subdir/file", []byte("smth"), 0664)
	afero.WriteFile(fs, "/dst/file", []byte("smth"), 0664)
	req := require.New(t)
	ctx := context.Background()
	err := SyncDirs(ctx, "/src", "/dst")
	req.NoError(err)
}

func TestScanSrc(t *testing.T) {
	req := require.New(t)
	fs.MkdirAll("/src/subdir", 0644)
	fs.Create("/src/file1")
	fs.Create("/src/subdir/file2")
	afero.WriteFile(fs, "/src/file1", []byte("smth"), 0664)
	afero.WriteFile(fs, "/src/subdir/file2", []byte("smth_else"), 0664)
	err := scanSrc("/src")
	req.NoError(err)
}

func TestScanDest(t *testing.T) {
	req := require.New(t)
	fs.MkdirAll("/src", 0644)
	SrcDir = "/src"
	fs.MkdirAll("/dst/subdir", 0664)
	fs.Create("/dst/subdir/file")
	afero.WriteFile(fs, "/dst/subdir/file", []byte("smth"), 0664)
	err := scanDest("/dst")
	req.NoError(err)
	filesToDelete = make(chan string, 10)
}

func TestSyncFiles(t *testing.T) {
	req := require.New(t)
	fs.MkdirAll("/src", 0664)
	fs.MkdirAll("/dst", 0664)
	fs.Create("/src/file")
	afero.WriteFile(fs, "/src/file", []byte("smth"), 0664)
	err := syncFiles("/src/file", "/dst/file")
	dstCont, _ := afero.ReadFile(fs, "/dst/file")
	req.Equal(string(dstCont), "smth")
	req.NoError(err)
}

func BenchmarkSyncFiles(b *testing.B) {
	fs.MkdirAll("/src", 0664)
	fs.MkdirAll("/dst", 0664)
	fs.Create("/src/file")
	afero.WriteFile(fs, "/src/file", []byte("smth"), 0664)
	for i := 0; i < b.N; i++ {
		syncFiles("/src/file", "/dst/file")
	}
}
