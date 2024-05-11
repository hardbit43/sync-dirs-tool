package synchronizer

import (
	"context"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

var (
	fs = afero.NewMemMapFs()
)

func TestSyncDirs(t *testing.T) {
	afero.WriteFile(fs, "/src/file1.txt", []byte("smth"), 0664)
	afero.WriteFile(fs, "/dst/file2.txt", []byte("smth"), 0664)
	req := require.New(t)
	ctx := context.Background()
	err := SyncDirs(ctx, "/src", "/dst")
	req.NoError(err)
}

func TestScanSrc(t *testing.T) {
	req := require.New(t)
	afero.WriteFile(fs, "/src/file1.txt", []byte("smth"), 0664)
	afero.WriteFile(fs, "/src/subdir/file2.txt", []byte("smth"), 0664)
	err := scanSrc("/src")
	req.NoError(err)
}

func TestScanDest(t *testing.T) {
	req := require.New(t)
	afero.WriteFile(fs, "/dst/file1.txt", []byte("smth"), 0664)
	afero.WriteFile(fs, "/dst/subdir/file2.txt", []byte("smth"), 0664)
	err := scanDest("/dst")
	req.NoError(err)
}

func TestSyncFiles(t *testing.T) {
	req := require.New(t)
	srcFile, _ := fs.Create("/src/file1.txt")
	dstFile, _ := fs.Create("/dst/file2.txt")
	afero.WriteFile(fs, "/src/file1.txt", []byte("smth"), 0664)
	afero.WriteFile(fs, "/dst/file2.txt", []byte("smth"), 0664)

	err := syncFiles(srcFile.Name(), dstFile.Name())
	req.NoError(err)
}
