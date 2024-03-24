package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
)

const (
	srcDir  = `C:\dev\project\simpleProgs\rebrain\finalTask\srcTestDir`
	destDir = `C:\dev\project\simpleProgs\rebrain\finalTask\testDir`
)

func main() {
	// defer fmt.Println("synchronized")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := syncDirs(ctx, srcDir, destDir); err != nil {
		log.Fatal(err)
	}
}

func syncDirs(ctx context.Context, srcDir, destDir string) error {
	filesToCopy := make(chan string)
	filesToDelete := make(chan string)
	filesToSync := make(chan string)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case file, ok := <-filesToCopy:
				if !ok {
					return
				}
				srcFilePath := filepath.Join(srcDir, file)
				dstFilePath := filepath.Join(destDir, file)
				if err := copyFile(srcFilePath, dstFilePath); err != nil {
					fmt.Printf("Failed to copy %s to %s: %v\n", srcFilePath, dstFilePath, err)
				}

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
				fmt.Println(synch)
				srcFilePath := filepath.Join(srcDir, synch)
				dstFilePath := filepath.Join(destDir, synch)
				if err := syncFiles(srcFilePath, dstFilePath); err != nil {
					fmt.Printf("Failed to synch %s to %s: %v\n", srcFilePath, dstFilePath, err)
				}

			case <-ctx.Done():
				return
			}
		}
	}()

	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(destDir, relPath)
		if _, err := os.Stat(dstPath); os.IsNotExist(err) {
			filesToCopy <- relPath // Файл отсутствует в целевой директории, добавляем в список для копирования.
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		} else {
			filesToSync <- relPath
		}

		dstPath := filepath.Join(destDir, relPath)
		_, err = os.Stat(dstPath)
		if err != nil {
			fmt.Println(err)
		}
		// fmt.Println(f)
		// Файл присутствует в целевой директории, добавляем в список для копирования.

		return nil
	})

	if err != nil {
		return fmt.Errorf("%v", err)
	}

	err = filepath.Walk(destDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(destDir, path)
		if err != nil {
			return err
		}

		srcPath := filepath.Join(srcDir, relPath)
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			filesToDelete <- relPath // Файл отсутствует в целевой директории, добавляем в список для копирования.
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	close(filesToCopy)
	close(filesToDelete)
	close(filesToSync)

	wg.Wait() // Ожидаем завершения работы горутины копирования.

	return nil
}

func copyFile(srcPath, destPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(srcFile, destFile)
	if err != nil {
		return err
	}

	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return err
	}

	return os.Chmod(destPath, srcInfo.Mode())
}

func syncFiles(srcPath, destPath string) error {
	var txt []byte

	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return err
	}

	err = os.Chmod(destPath, srcInfo.Mode())
	if err != nil {
		return err
	}

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.OpenFile(destPath, os.O_APPEND, 0664)
	if err != nil {
		return err
	}
	defer destFile.Close()

	err = destFile.Truncate(0)
	if err != nil {
		fmt.Println(err)
	}

	reader := bufio.NewReader(srcFile)
	reader.Read(txt)
	_, err = destFile.Write(txt)
	if err != nil {
		fmt.Println(err)
	}
	return nil
}
