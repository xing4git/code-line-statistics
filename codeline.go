// a simple go program for computing total line of souce files stored in one dir
package main

import (
	"fmt"
	"bufio"
	"os"
	"sync"
	"strings"
)

var (
	linesum int
	mutex   *sync.Mutex = new(sync.Mutex)
)

var (
	// the dir where souce file stored
	rootPath string = "/home/xing/source/go/src"
	// exclude these sub dirs
	nodirs [4]string = [4]string{"/bitbucket.org", "/github.com", "/goplayer", "/uniqush"}
	// the suffix name you care
	suffixname string = ".go"
)

func main() {
	// sync chan using for waiting
	done := make(chan bool)
	go codeLineSum(rootPath, done)
	<-done

	fmt.Println("total line:", linesum)
}

// compute souce file line number
func codeLineSum(root string, done chan bool) {
	var goes int              // children goroutines number
	godone := make(chan bool) // sync chan using for waiting all his children goroutines finished
	isDstDir := checkDir(root)
	defer func() {
		if pan := recover(); pan != nil {
			fmt.Printf("root: %s, panic:%#v\n", root, pan)
		}

		// waiting for his children done
		for i := 0; i < goes; i++ {
			<-godone
		}

		if isDstDir {
			fmt.Printf("dir  %s complete\n", root)
		}
		// this goroutine done, notify his parent
		done <- true
	}()
	if !isDstDir {
		return
	}

	rootfi, err := os.Lstat(root)
	checkerr(err)

	rootdir, err := os.Open(root)
	checkerr(err)

	if rootfi.IsDir() {
		fis, err := rootdir.Readdir(0)
		checkerr(err)
		goes = len(fis) // current goroutine has len(fis) children
		for _, fi := range fis {
			if fi.IsDir() {
				go codeLineSum(root+"/"+fi.Name(), godone)
			} else {
				go readfile(root+"/"+fi.Name(), godone)
			}
		}
	} else {
		goes = 1 // if rootfi is a file, current goroutine has only one child
		go readfile(root, godone)
	}
}

func readfile(filename string, done chan bool) {
	var line int
	isDstFile := strings.HasSuffix(filename, suffixname)
	defer func() {
		if pan := recover(); pan != nil {
			fmt.Printf("filename: %s, panic:%#v\n", filename, pan)
		}
		if isDstFile {
			addLineNum(line)
			fmt.Printf("file %s complete, line = %d\n", filename, line)
		}
		// this goroutine done, notify his parent
		done <- true
	}()
	if !isDstFile {
		return
	}

	file, err := os.Open(filename)
	defer file.Close()
	checkerr(err)

	reader := bufio.NewReader(file)
	for {
		_, isPrefix, err := reader.ReadLine()
		if err != nil {
			break
		}
		if !isPrefix {
			line++
		}
	}
}

// check whether this dir is the dest dir
func checkDir(dirpath string) bool {
	for _, dir := range nodirs {
		if rootPath+dir == dirpath {
			return false
		}
	}
	return true
}

func addLineNum(num int) {
	mutex.Lock()
	defer mutex.Unlock()
	linesum += num
}

// if error happened, throw a panic, and the panic will be recover in defer function
func checkerr(err error) {
	if err != nil {
		panic(err.Error())
	}
}
