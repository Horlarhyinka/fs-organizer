package worker

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
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

var (
	OutputBaseExt = "ext"
	OutputBaseDate = "date"
)

const DefaultBaseMapping = "unknown"

type Option struct {
	ConflictOpt string
	DryRun	bool
	MoveFiles	bool
	Exclude  []string
	Mappings map[string]string
	OutputBase string
}

func validateDir(p string) error {
	info, err := os.Lstat(p)
	if err != nil {
		return fmt.Errorf("invalid filepath %q", p)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("invalid filepath %q - symlink detected", p)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a valid directory", p)
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

func pathMatch(p string, m []string) bool {
	for _, v := range m {
		match, err := filepath.Match(v, p)
		if err != nil {
			fmt.Println("invalid file pattern", v, "match skipped for", p)
			continue
		}
		if match {
			return true
		}
	}
	return false
}

func getDirectoryNodes(p string, excludes []string) ([]FileNode, error) {
	nodes := make([]FileNode, 0)
	var exec func(fp string)
	errs := make([]error, 0)

	exec = func(fp string) {
		if pathMatch(fp, excludes) {
			fmt.Printf("%s excluded, skipping...\n", fp)
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

	f := &FileNode{
		ext:       filepath.Ext(p),
		timestamp: file.ModTime(),
		name:      file.Name(),
		fullpath:  p,
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

func getShiftEntries(nodes []FileNode, out string, base string, mapping map[string]string, onConflict string) ([]ShiftEntry, int) {
	stateMap := make(map[string]string)
	res := make([]ShiftEntry, 0)
	destMem := make(map[string]int, 0)
	var conflicts int
	re := regexp.MustCompile(`^(.+)(\.[^.]+)`)
	for _, node := range nodes {
		key := ""
		if base == OutputBaseExt {
			m, ok := mapping[strings.ToLower(node.ext)]
			if ok {
				key = m
			}else{
				key = strings.ToLower(node.ext)
			}
		}
		if base == OutputBaseDate {
			key = strings.Split(node.timestamp.String(), " ")[0]
		}
		if key == "" {
			key = DefaultBaseMapping
		}
		if _, ok := stateMap[key]; !ok {
			stateMap[key] = path.Join(out, key)
		}
		destPth := path.Join(stateMap[key], node.name)
		existing := destMem[destPth]
		if existing > 0 {
			switch onConflict {
			case ConflictOptSkip:
				continue
			case ConflictOptRename:
				destPth = re.ReplaceAllString(destPth, fmt.Sprintf(`${1}(%d)${2}`, existing))
				fmt.Println("renaming on conflict", destPth)
			default:
			}
			conflicts++
		}
		res = append(res, ShiftEntry{From: node.fullpath, To: destPth})
		destMem[destPth] += 1
	}
	return res, conflicts
}


func logEntries(entries []ShiftEntry) {
	for i, en := range entries {
		fmt.Printf("%d. %s => %s\n", i+1, en.From, en.To)
	}
}

func shiftFiles(_ context.Context, entries []ShiftEntry, deleteSource bool) error {
	var failed int
	var errGrp error
	for _, se := range entries {
		if err := moveFile(se.From, se.To, deleteSource); err != nil {
			errGrp = errors.Join(errGrp, err)
			failed++
		}
	}
	fmt.Printf("processed %d of %d files\n", len(entries) - failed, len(entries))
	if errGrp != nil {
		fmt.Println(errGrp)
	}
	return nil
}

func moveFile(from, to string, deleteSource bool) error {
	cnt, err := os.ReadFile(from); if err != nil {
		return err
	}
	destDir := filepath.Dir(to)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	if err := os.WriteFile(to, cnt, 0644); err != nil {
		return err
	}
	if deleteSource {
		err = os.Remove(from); if err != nil {
			fmt.Printf("failed to remove file %s - default to copy mode: %v", from, err)
		}
	}
	return nil
}


func Organize(in string, out string, opts Option) error {
	fmt.Println("organizing", in)
	//validate path is valid and exists
	if err := validateDir(in); err != nil {
		return fmt.Errorf("invalid input path %s: %w", in, err)
	}

	//construct FileNode list from the current tree
	nodes, err := getDirectoryNodes(in, opts.Exclude); if err != nil {
		return err
	}
	//get shift entries
	entries, conflicts := getShiftEntries(nodes, out, opts.OutputBase, opts.Mappings, opts.ConflictOpt)
	if opts.DryRun {
		logEntries(entries)
		fmt.Printf("Total: %d, Conflicts: %d (on-conflict = %s)", len(entries), conflicts, opts.ConflictOpt)
		return nil
	}
	return shiftFiles(context.Background(), entries, opts.MoveFiles)
}
