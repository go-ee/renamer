package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var lPath = flag.String("l", "", "left folder")
var rPath = flag.String("r", "", "rigth folder")

var folder = flag.String("folder", "", "folder with files")
var search = flag.String("search", "", "search 'from' string")
var replace = flag.String("replace", "", "replace with 'to' string")

var tPath = flag.String("t", "", "target folder for renamed files")

var ext = flag.String("ext", ".jpg", "file extension, e.g. jpg")

func main() {
	flag.Parse()
	if *lPath != "" && *rPath != "" && *tPath != "" {
		combineLeftRigth(*lPath, *rPath, *tPath)
	} else if *folder != "" && *search != "" && *replace != "" && *tPath != "" {
		rename(*folder, *search, *replace, *tPath)
	} else {
		flag.Usage()
	}
}

func rename(folder string, search string, replace string, target string) {
	fmt.Printf("rename search '%v' with ', '%v' in %v to %v\n", search, replace, folder, target)
	files, err := readDir(folder, *ext)
	if err != nil {
		log.Fatal(err)
	}

	var fileName string
	for _, file := range files {
		fileName = strings.Replace(file.Name(), search, replace, -1)
		//fmt.Printf("%v, %v\n", file.Name(), fileName)
		copyFile(filepath.Join(folder, file.Name()), filepath.Join(target, fileName))
	}
}

func combineLeftRigth(lPath string, rPath string, target string) {
	lFiles, err := readDir(lPath, *ext)
	if err != nil {
		log.Fatal(err)
	}

	rFiles, err := readDir(rPath, *ext)
	if err != nil {
		log.Fatal(err)
	}

	if len(lFiles) != len(rFiles) {
		log.Fatalf("Different files count in left (%v) and right (%v) folders", len(lFiles), len(rFiles))
	}

	padLength := len(fmt.Sprintf("%v", len(lFiles)*2))

	for i := len(rFiles)/2 - 1; i >= 0; i-- {
		opp := len(rFiles) - 1 - i
		rFiles[i], rFiles[opp] = rFiles[opp], rFiles[i]
	}

	fmt.Printf("merge from %v and reversed %v to %v\n", lPath, rPath, target)
	fileIndex := 0

	var fileName string

	for i, l := range lFiles {
		fileIndex = fileIndex + 1
		fileName = padLeft(strconv.Itoa(fileIndex), "0", padLength) + "-" + l.Name()
		//fmt.Printf("%v, %v\n", fileIndex, fileName)
		copyFile(filepath.Join(lPath, l.Name()), filepath.Join(target, fileName))

		r := rFiles[i]

		fileIndex = fileIndex + 1

		fileName = padLeft(strconv.Itoa(fileIndex), "0", padLength) + "-" + r.Name()
		//fmt.Printf("%v, %v\n", fileIndex, fileName)
		copyFile(filepath.Join(rPath, r.Name()), filepath.Join(target, fileName))
	}
}

func visitLog(path string, f os.FileInfo, err error) error {
	fmt.Println(path)
	return nil
}

func visitRename(path string, f os.FileInfo, err error) error {
	if name := f.Name(); strings.HasPrefix(name, "name_") {
		dir := filepath.Dir(path)
		newname := strings.Replace(name, "name_", "name1_", 1)
		newpath := filepath.Join(dir, newname)
		fmt.Printf("mv %q %q\n", path, newpath)
		os.Rename(path, newpath)
	}
	return nil
}

// CopyFile copies a file from src to dst. If src and dst files exist, and are
// the same, then return success. Otherise, attempt to create a hard link
// between the two files. If that fail, copy the file contents from src to dst.
func copyFile(src, dst string) (err error) {
	sfi, err := os.Stat(src)
	if err != nil {
		return
	}
	if !sfi.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
		if os.SameFile(sfi, dfi) {
			return
		}
	}
	if err = os.Link(src, dst); err == nil {
		return
	}
	err = copyFileContents(src, dst)
	return
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

func readDir(path string, ext string) ([]os.FileInfo, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	ret := files[:0]
	for _, file := range files {
		if strings.EqualFold(filepath.Ext(file.Name()), ext) {
			ret = append(ret, file)
		}
	}
	fmt.Printf("%v, %v files\n", path, len(ret))
	return ret, nil
}

func padLeft(str, pad string, lenght int) string {
	for {
		str = pad + str
		if len(str) > lenght {
			return str[1 : lenght+1]
		}
	}
}
