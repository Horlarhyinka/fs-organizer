package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"
	"time"
)

type FileNode struct {
	ext       string
	timestamp time.Time
	name      string
	fullpath  string
}

func validateDir(p string) error {
	info, err := os.Lstat(p)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("invalid filepath %q", p)
	}
	if info.Mode()&os.ModeSymlink == 1 {
		return fmt.Errorf("invalid filepath %q - symlink detected", p)
	}
	return nil
}

func readDirChildren(p string) ([]string, error) {
	fsys := os.DirFS(p)
	entries, err := fs.ReadDir(fsys, p); if err != nil {
		return nil, err
	}
	children := make([]string, 0)
	for _, d := range entries {
		children = append(children, path.Join(p, d.Name()))
	}
	return children, err
}

func getDirectoryNodes(p string, excludes []string) ([]FileNode, error) {
	nodes := make([]FileNode, 0)
	var exec func(fp string)
	errs := make([]error, 0)
	excState := make(map[string]bool)
	for _, v := range excludes {
		excState[v] = true
	}
	exec = func(fp string) {
		if excState[fp] {
			return
		}
		if isFileErr := isValidFile(fp); isFileErr == nil {
			fn, fnErr := fileToNode(fp); if fnErr != nil {
				errs = append(errs, fnErr)
				return
			}
			nodes = append(nodes, *fn)
			return
		}
		isDirErr := validateDir(fp); if isDirErr != nil {
			errs = append(errs, isDirErr)
			return 
		}
		children, err := readDirChildren(fp); if err != nil {
			errs = append(errs, err)
			return
		}
		fmt.Printf("children of %s are %s\n", fp, strings.Join(children, ", "))
		for _, c := range children {
			exec(path.Join(fp, c))
		}
		
	}
	exec(p)
	var longErr error
	if len(errs) > 0 {
		for _, e := range errs {
			longErr = errors.Join(e, longErr)
		}
		return nil, longErr
	}
	return nodes, nil
}


func fileToNode(p string) (*FileNode, error) {
	if err := isValidFile(p); err != nil {
		return nil, err
	}
	file, err := os.Stat(p)
	if err != nil {
		return nil, err
	}
	f := &FileNode{
		ext:       "",
		timestamp: file.ModTime(),
		name:      file.Name(),
		fullpath:  p,
	}
	seperated := strings.Split(p, ".")
	if len(seperated) > 0 {
		f.ext = seperated[len(seperated)-1]
	}
	return f, nil
}

func isValidFile(p string) error {
	info, err := os.Stat(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("invalid file: %q does not exist", p)
		}
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("invalid file: %q is a directory", p)
	}
	return nil
}

func OrganizeNodes(nodes []FileNode, out string) []FileNode {
	stateMap := make(map[string]string)
	res := make([]FileNode, 0)
	for _, node := range nodes {
		if _, ok := stateMap[node.ext]; !ok {
			stateMap[node.ext] = path.Join(out, node.ext)
		}
		node.fullpath = path.Join(stateMap[node.ext], node.name)
		res = append(res, node)
	}
	return res
}

func RunWorker(fp, mode string, excludes []string) error {
	//validate path is valid and exists
	if err := validateDir(fp); err != nil {
		return err
	}
	//construct FileNode list from the current tree
	nodes, nodesErr := getDirectoryNodes(fp, excludes); if nodesErr != nil {
		return nodesErr
	}
	//run through categorization engine
	OrganizeNodes(nodes, "out")
	//categorization engine returns a new fs-tree

	//copy or move directories to match constructed file tree

	return nil
}
