package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"
)

// FileManager implements file manager window
type FileManager struct {
	window         *walk.MainWindow
	fileFilterEdit *walk.LineEdit
	dirLabel       *walk.Label
	filesListBox   *walk.ListBox

	// current directory
	dir string
	// files in a current directory
	files []string
	// string by which we filter files
	fileFilter string
	// files after being filtered by fileFilter
	filteredFiles []string
}

// must be called synchronized
func (fm *FileManager) setFiles(dir string, files []string) {
	fm.files = files
	s := fmt.Sprintf("Directory: %s", dir)
	fm.dirLabel.SetText(s)
	fm.dir = dir
	fm.applyFileFilter(fm.fileFilter)
}

func getFilteredFiles(files []string, filter string) []string {
	if filter == "" {
		return files
	}
	var filtered []string
	filter = strings.ToLower(filter)
	for _, f := range files {
		f = strings.ToLower(f)
		if strings.Contains(f, filter) {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

// must be called synchronized
func (fm *FileManager) applyFileFilter(filter string) {
	fm.fileFilter = filter
	fm.filteredFiles = getFilteredFiles(fm.files, fm.fileFilter)
	fm.filesListBox.SetModel(fm.filteredFiles)
}

// must be called synchronized
func (fm *FileManager) onFilterChanged() {
	filter := fm.fileFilterEdit.Text()
	fm.applyFileFilter(filter)
}

// must be called synchronized
func (fm *FileManager) onFileClicked() {
	idx := fm.filesListBox.CurrentIndex()
	if idx < 0 || idx > len(fm.files) {
		must(errors.New("invalid index"))
	}
	file := fm.files[idx]
	isDir := file == ".." || strings.HasSuffix(file, "/")
	if !isDir {
		return
	}
	dir := filepath.Join(fm.dir, file)
	fm.showFiles(dir)
}

func (fm *FileManager) showFiles(dir string) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		s := fmt.Sprintf("Invalid directory %s", dir)
		fm.dirLabel.SetText(s)
		return
	}
	s := fmt.Sprintf("Scanning directory %s", dir)
	fm.dirLabel.SetText(s)
	fm.filesListBox.SetModel([]string{})
	go func() {
		osfiles, err := ioutil.ReadDir(dir)
		if err != nil {
			fm.dirLabel.Synchronize(func() {
				s := fmt.Sprintf("Error reading directory %s, error: %s", dir, err.Error())
				fm.dirLabel.SetText(s)
			})
			return
		}
		var files []string
		if filepath.Dir(dir) != dir {
			files = append(files, "..")
		}
		for _, file := range osfiles {
			s := file.Name()
			if file.IsDir() {
				s = s + "/"
			}
			files = append(files, s)
		}
		fm.window.Synchronize(func() {
			fm.setFiles(dir, files)
		})
	}()
}

func createFileManager(startDir string) (*FileManager, error) {
	var fm FileManager
	def := declarative.MainWindow{
		AssignTo: &fm.window,
		Title:    "File Manageer",
		MinSize:  declarative.Size{Width: 240, Height: 200},
		Size:     declarative.Size{Width: 320, Height: 400},
		Layout:   declarative.VBox{},
		Children: []declarative.Widget{
			declarative.LineEdit{
				AssignTo:      &fm.fileFilterEdit,
				Visible:       true,
				CueBanner:     "Enter file filter",
				OnTextChanged: fm.onFilterChanged,
				OnKeyUp: func(key walk.Key) {
					if key == walk.KeyEscape {
						fm.fileFilterEdit.SetText("")
					}
				},
			},
			declarative.Label{
				AssignTo: &fm.dirLabel,
				Visible:  true,
				Text:     "Directory:",
			},
			declarative.ListBox{
				AssignTo:        &fm.filesListBox,
				Visible:         true,
				OnItemActivated: fm.onFileClicked,
			},
		},
	}
	err := def.Create()
	if err != nil {
		return nil, err
	}
	fm.showFiles(startDir)
	return &fm, nil
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func doFileManager() {
	fm, err := createFileManager(".")
	if err != nil {
		log.Fatalf("createFileManager failed with '%s'\n", err)
	}
	_ = fm.window.Run()
}
