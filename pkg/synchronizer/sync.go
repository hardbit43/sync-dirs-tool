package synchronizer

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/dustin/go-humanize"
)

const (
	SrcDir  = "./srcTestDir"
	DestDir = "./testDir"
)

var (
	filesToDelete = make(chan string)
	filesToSync   = make(chan string)
)

func scanSrc(SrcDir string) error {
	filepath.Walk(SrcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(SrcDir, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		dstPath := filepath.Join(DestDir, relPath)

		if info.IsDir() {
			slog.Info(relPath, "synchronizing directory size is", humanize.Bytes(uint64(info.Size())))
			return os.MkdirAll(dstPath, info.Mode())
		} else {
			filesToSync <- relPath
			slog.Info(relPath, "synchronizing file size is", humanize.Bytes(uint64(info.Size())))
		}

		return nil

	})
	return nil
}

func scanDest(DestDir string) error {
	filepath.Walk(DestDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(DestDir, path)
		if err != nil {
			return err
		}

		srcPath := filepath.Join(SrcDir, relPath)
		_, err = os.Stat(srcPath)
		if os.IsNotExist(err) {
			slog.Info(relPath, "deleting file size is", humanize.Bytes(uint64(info.Size())))
			filesToDelete <- relPath
		}

		return nil
	})
	return nil
}

func syncFiles(srcPath, destPath string) error {
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return err
	}

	srcFile, err := os.OpenFile(srcPath, os.O_CREATE|os.O_RDWR, 0664)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY, srcInfo.Mode())
	if err != nil {
		return err
	}

	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}

	err = os.Chmod(destPath, srcInfo.Mode())
	if err != nil {
		return err
	}

	return nil
}

func SyncDirs(ctx context.Context, SrcDir, DestDir string) error {

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {

			case del, ok := <-filesToDelete:
				if !ok {
					return
				}
				dstFilePath := filepath.Join(DestDir, del)
				err := os.RemoveAll(dstFilePath)
				if err != nil {
					slog.Error("error delete files due to")
				}

			case synch, ok := <-filesToSync:
				if !ok {
					return
				}
				srcFilePath := filepath.Join(SrcDir, synch)
				dstFilePath := filepath.Join(DestDir, synch)
				err := syncFiles(srcFilePath, dstFilePath)
				if err != nil {
					slog.Error("error sync files due to", err)
					return
				}

			case <-ctx.Done():
				return
			}
		}
	}()

	err := scanSrc(SrcDir)
	if err != nil {
		slog.Error("error making files to sync", err)
	}

	err = scanDest(DestDir)

	if err != nil {
		slog.Error("error making file to delete", err)
	}

	close(filesToSync)
	close(filesToDelete)
	wg.Wait()
	return nil
}
