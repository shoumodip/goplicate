package main

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func hashFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return string(hash.Sum(nil)), nil
}

type Walker struct {
	base  map[string]string
	extra []string
}

func newWalker() Walker {
	return Walker{
		base:  map[string]string{},
		extra: []string{},
	}
}

func (w *Walker) add(path string) error {
	items, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, item := range items {
		name := filepath.Join(path, item.Name())
		if item.IsDir() {
			if err = w.add(name); err != nil {
				return err
			}
		} else {
			hash, err := hashFile(name)
			if err != nil {
				return err
			}

			if _, ok := w.base[hash]; ok {
				fmt.Println("Found", name)
				w.extra = append(w.extra, name)
			} else {
				w.base[hash] = name
			}
		}
	}

	return nil
}

func (w *Walker) save(path string) error {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(path, "list.txt"))
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for i, item := range w.extra {
		err = os.Rename(item, filepath.Join(path, fmt.Sprint(i)))
		if err != nil {
			return err
		}
		fmt.Fprintln(writer, item)
	}

	return nil
}

func usage() {
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  "+os.Args[0], "MODE PATH")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Modes:")
	fmt.Fprintln(os.Stderr, "  gather   Gather duplicates into PATH")
	fmt.Fprintln(os.Stderr, "  restore  Restore gathered duplicates from PATH")
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
		usage()
		os.Exit(1)
	}

	mode := os.Args[1]
	switch mode {
	case "gather":
		base := os.Args[2]
		walker := newWalker()
		handleError(walker.add("."))
		handleError(walker.save(base))
		fmt.Println("Gathered", len(walker.extra), "duplicates into '"+base+"'")

	case "restore":
		base := os.Args[2]

		data, err := os.ReadFile(filepath.Join(base, "list.txt"))
		handleError(err)

		list := strings.Split(string(data), "\n")
		for i, path := range list {
			if path == "" {
				continue
			}
			handleError(os.Rename(filepath.Join(base, fmt.Sprint(i)), path))
		}

		handleError(os.RemoveAll(base))
		fmt.Println("Extracted", len(list)-1, "duplicates from '"+base+"'")

	default:
		fmt.Fprintln(os.Stderr, "Error: invalid mode '"+mode+"'")
		usage()
	}
}
