//
// lsof list opened file+pid
// support linux only
// Modified from wheelcomplex's lsof library
// https://github.com/wheelcomplex/lsof
//

package lsof

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// PIDs list PIDs opened
type PIDs map[int]struct{}

// InfoList list all opened files
type InfoList struct {
	files map[string]PIDs // list index by file
}

// newInfoList
func newInfoList() *InfoList {
	l := new(InfoList)
	l.files = make(map[string]PIDs)
	return l
}

// Open start a goroutine and read
func Open(listdir string, prefix string) (*InfoList, error) {
	var err error
	if len(listdir) == 0 {
		listdir = "/proc/"
	}
	if prefix == "" {
		prefix = "."
	}
	prefix = filepath.Clean(prefix)
	listdir, err = filepath.Abs(listdir)
	if err != nil {
		return nil, err
	}
	l := newInfoList()
	err = l.readDir(listdir, prefix)

	if err != nil {
		fmt.Printf("readDir %s: %s\n", listdir, err.Error())
	}
	return l, err
}

// Lsof start a goroutine and read File2PIDs send to output channel
func Lsof(prefix string) (*InfoList, error) {
	return Open("", prefix)
}

// readPIDDir
func (l *InfoList) readPIDDir(onepid int, fddir string, prefix string) error {
	var err error
	var symlist []string
	var onefile string
	var checkprefix bool = true
	if prefix == "." {
		checkprefix = false
	}
	symlist, err = filepath.Glob(fddir)
	if err != nil {
		//fmt.Printf("Error: Glob %s: %s\n", fddir, err.Error())
		return err
	}
	if len(symlist) == 0 {
		return nil
	}
	for _, onefd := range symlist {
		onefile, err = filepath.EvalSymlinks(onefd)
		if err != nil {
			//fmt.Printf("Error: EvalSymlinks %s: %s\n", onefd, err.Error())
			continue
		}
		if checkprefix {
			if strings.HasPrefix(onefile, prefix) == false {
				//fmt.Printf("Skip: pid %d open fd %s -> %s\n", onepid, onefd, onefile)
				continue
			}
		}
		//fmt.Printf("Got: pid %d open fd %s -> %s\n", onepid, onefd, onefile)
		if _, ok := l.files[onefile]; !ok {
			pidMap := make(PIDs)
			l.files[onefile] = pidMap
		}
		l.files[onefile][onepid] = struct{}{}
	}
	return nil
}

// readDir
func (l *InfoList) readDir(path string, prefix string) error {
	var err error
	var cwd string

	cwd, err = os.Getwd()
	if err != nil {
		fmt.Printf("Getwd: %s\n", err.Error())
		cwd = "/"
	}
	defer func() {
		err := os.Chdir(cwd)
		if err != nil {
			fmt.Printf("WARNING: Chdir %s: %s\n", cwd, err.Error())
		}
	}()
	err = os.Chdir(path)
	if err != nil {
		fmt.Printf("Error: Chdir %s: %s\n", path, err.Error())
		return err
	}
	// list pid dir in current dir
	var pidlist []string
	pidlist, err = filepath.Glob("*")
	if err != nil {
		fmt.Printf("Error: Glob %s: %s\n", path, err.Error())
		return err
	}
	for _, val := range pidlist {
		onepid, err := strconv.Atoi(val)
		if err != nil {
			// no a pid-dir
			//fmt.Printf("Error: Atoi %s: %s\n", path+"/"+val, err.Error())
			continue
		}
		fddir := path + "/" + val + "/fd/*"
		//func (l *InfoList) readPIDDir(onepid int, fddir string, prefix string) error
		err = l.readPIDDir(onepid, fddir, prefix)
		if err != nil {
			fmt.Printf("Error: readPIDDir %s: %s\n", fddir, err.Error())
			continue
		}
	}
	return nil
}

// GetFDCountForFile return file list in map
func (l *InfoList) GetFDCountForFile(fileLoc string) (int, error) {
	res, ok := l.files[fileLoc]
	if !ok {
		return 0, ErrFileDescriptorsNotFound
	}
	return len(res), nil
}

// File2PIDsMap return file list in map
func (l *InfoList) File2PIDsMap(fileLoc string) map[string]PIDs {
	return l.files
}

