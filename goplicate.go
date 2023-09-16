package main

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const MaxHashBytes = 4096

func hashFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err = io.CopyN(hash, file, MaxHashBytes); err != nil && err != io.EOF {
		return "", err
	}

	return string(hash.Sum(nil)), nil
}

type File struct {
	path string
	hash string
}

func gatherDuplicates(dir string) error {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if strings.HasPrefix(dir, cwd) {
		return errors.New("cannot gather duplicates into directory inside current directory")
	}

	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	list := map[int64][]File{}
	count := 0

	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			size := info.Size()
			if prev, ok := list[size]; ok {
				hash, err := hashFile(path)
				if err != nil {
					return err
				}

				for _, item := range prev {
					if item.hash == "" {
						item.hash, err = hashFile(item.path)
						if err != nil {
							return err
						}
					}

					if item.hash == hash {
						prev, err := os.ReadFile(item.path)
						if err != nil {
							return err
						}

						this, err := os.ReadFile(path)
						if err != nil {
							return err
						}

						if bytes.Equal(prev, this) {
							fmt.Println("Found", path)

							abs, err := filepath.Abs(path)
							if err != nil {
								return err
							}

							err = os.Symlink(abs, filepath.Join(dir, fmt.Sprint(count)))
							if err != nil {
								return err
							}

							count++
						}
					}
				}
			}

			list[size] = append(list[size], File{path: path, hash: ""})
		}

		return nil
	})

	if err != nil {
		return err
	}

	fmt.Println("Gathered", count, "duplicates into '"+dir+"'")
	return nil
}

func deleteDuplicates(dir string) error {
	items, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, item := range items {
		path, err := os.Readlink(filepath.Join(dir, item.Name()))
		if err != nil {
			return err
		}

		err = os.Remove(path)
		if err != nil {
			return err
		}
	}

	err = os.RemoveAll(dir)
	if err != nil {
		return err
	}

	fmt.Println("Deleted", len(items), "duplicates from '"+dir+"'")
	return nil
}

func showUsage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  goplicate MODE DIR")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Modes:")
	fmt.Fprintln(os.Stderr, "  gather  Gather duplicates into DIR")
	fmt.Fprintln(os.Stderr, "  delete  Delete gathered duplicates in DIR")
	os.Exit(1)
}

func handleError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Error: insufficient arguments")
		fmt.Fprintln(os.Stderr)
		showUsage()
	}

	switch mode := os.Args[1]; mode {
	case "gather":
		handleError(gatherDuplicates(os.Args[2]))

	case "delete":
		handleError(deleteDuplicates(os.Args[2]))

	default:
		fmt.Fprintln(os.Stderr, "Error: invalid mode '"+mode+"'")
		fmt.Fprintln(os.Stderr)
		showUsage()
	}
}
