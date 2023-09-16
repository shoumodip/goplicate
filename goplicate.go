package main

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	list := map[int64][]File{}
	items := []string{}

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
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
							items = append(items, path)
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

	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	for i, item := range items {
		abs, err := filepath.Abs(item)
		if err != nil {
			return err
		}

		err = os.Symlink(abs, filepath.Join(dir, fmt.Sprint(i)))
		if err != nil {
			return err
		}
	}

	fmt.Println("Gathered", len(items), "duplicates into '"+dir+"'")
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
