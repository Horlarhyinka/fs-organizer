package worker

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

type ShiftEntry struct {
	From string
	To	 string
}

var (
	ConflictOptSkip = "skip"
	ConflictOptRename = "rename"
	ConflictOptOverwrite = "overwrite"
)


type Option struct {
	ConflictOpt string
	DryRun	bool
	MoveFiles	bool
	Exclude  []string
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
	entries, err := fs.ReadDir(fsys, "."); if err != nil {
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
				fmt.Println("invalid file:", fp)
				errs = append(errs, fnErr, fnErr)
				return
			}
			nodes = append(nodes, *fn)
			return
		}
		isDirErr := validateDir(fp); if isDirErr != nil {
			fmt.Println("invalid directory:", fp, isDirErr)
			errs = append(errs, isDirErr)
			return 
		}
		children, err := readDirChildren(fp); if err != nil {
			fmt.Println("unable to read children from", fp, err)
			errs = append(errs, err)
			return
		}
		for _, c := range children {
			exec(c)
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
	var ext string
	spl := strings.Split(file.Name(), ".")
	if len(spl) > 1 {
		ext = spl[len(spl) - 1]
	}
	f := &FileNode{
		ext:       ext,
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

func OrganizeNodes(nodes []FileNode, out string) []ShiftEntry {
	stateMap := make(map[string]string)
	res := make([]ShiftEntry, 0)
	for _, node := range nodes {
		if _, ok := stateMap[node.ext]; !ok {
			stateMap[node.ext] = path.Join(out, node.ext)
		}
		res = append(res, ShiftEntry{From: node.fullpath, To: path.Join(stateMap[node.ext], node.name)})
	}
	return res
}

func ShiftFiles(entries []ShiftEntry, mode string) error {
	// will use mode to determine if it drops old copy or not later. Just log for now
	fmt.Printf("sorting files in %s mode", mode)
	for _, en := range entries {
		fmt.Println(en.From, "->", en.To)
		content, err := os.ReadFile(en.From); if err != nil {
			return err
		}
		if err := os.WriteFile(en.To, content, 0755); err != nil {
			return err
		}
	}
	return nil
}

func RunWorker(fp, mode string, excludes []string) error {
	//validate path is valid and exists
	if err := validateDir(fp); err != nil {
		fmt.Println("invalid directiry:", fp)
		return err
	}
	//construct FileNode list from the current tree
	nodes, nodesErr := getDirectoryNodes(fp, excludes); if nodesErr != nil {
		fmt.Println("failed to get tree nodes from", fp)
		return nodesErr
	}
	//run through categorization engine
	sens := OrganizeNodes(nodes, "out")
	//categorization engine returns a new fs-tree
	ShiftFiles(sens, "w")
	//copy or move directories to match constructed file tree

	return nil
}

func Organize(in string, out string, opts Option) error {

	return nil
}
