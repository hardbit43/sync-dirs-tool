package main

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

var (
	filesToDelete = make(chan string)
	filesToSync   = make(chan string)
)

const (
	srcDir  = "./srcTestDir"
	destDir = "./testDir"
)

func main() {
	initLogger()
	defer slog.Info("synchronized")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := syncDirs(ctx, srcDir, destDir); err != nil {
		slog.Error(err.Error())
	}
}

func initLogger() {
	file, err := os.OpenFile("log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {

	}
	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}
	logger := slog.New(slog.NewJSONHandler(file, opts))
	slog.SetDefault(logger)
}

func scanSrc(srcDir string) error {
	filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		dstPath := filepath.Join(destDir, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		if !info.IsDir() {
			filesToSync <- relPath
		}
		slog.Info("size of %s is %d bytes", relPath, info.Size())
		return nil

	})
	return nil
}

func scanDest(destDir string) error {
	filepath.Walk(destDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(destDir, path)
		if err != nil {
			return err
		}

		srcPath := filepath.Join(srcDir, relPath)
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			filesToDelete <- relPath
		}

		return nil
	})
	return nil
}

func syncFiles(srcPath, destPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0664)
	if err != nil {
		return err
	}

	defer destFile.Close()

	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return err
	}

	err = os.Chmod(destPath, srcInfo.Mode())
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(srcFile)
	for scanner.Scan() {
		text := scanner.Text()
		_, err = destFile.Write([]byte(text))
		if err != nil {
			return err
		}
	}

	return nil
}

func syncDirs(ctx context.Context, srcDir, destDir string) error {
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
				dstFilePath := filepath.Join(destDir, del)
				os.Remove(dstFilePath)

			case synch, ok := <-filesToSync:
				if !ok {
					return
				}
				srcFilePath := filepath.Join(srcDir, synch)
				dstFilePath := filepath.Join(destDir, synch)
				err := syncFiles(srcFilePath, dstFilePath)
				if err != nil {
					fmt.Printf("Failed to synch %s with %s: %v\n", srcFilePath, dstFilePath, err)
				}

			case <-ctx.Done():
				return
			}
		}
	}()

	err := scanSrc(srcDir)

	if err != nil {
		return fmt.Errorf("error making files to synch %v", err)
	}

	err = scanDest(destDir)

	if err != nil {
		return fmt.Errorf("error making files to delete %v", err)
	}

	close(filesToDelete)
	close(filesToSync)
	wg.Wait()
	return nil
}
