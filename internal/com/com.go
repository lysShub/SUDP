package com

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

// Writers file handles
var Writers []io.Writer

func init() {
	Writers = append(Writers, os.Stderr)
}

// Errorlog logger
func Errorlog(err ...error) bool {
	// writers = []io.Writer{
	// 	errLogHandle,
	// 	os.Stdout,
	// }
	var haveErr bool = false
	for i, e := range err {
		if e != nil {
			haveErr = true
			_, fp, ln, _ := runtime.Caller(1) //行数

			w := io.MultiWriter(Writers...)
			logger := log.New(w, "", log.Ldate|log.Ltime) //|log.Lshortfile
			logger.Println(fp + ":" + strconv.Itoa(ln) + "." + strconv.Itoa(i+1) + "==>" + e.Error())
		}
	}
	return haveErr
}

// Info floder info
type Info struct {
	S []int64  //size 字节
	N []string //name 相对路径
	T []int64  //time 纳秒
}

// GetFloderInfo 获取文件夹信息
// return: Info []string(非正常文件)
func GetFloderInfo(path string) (Info, string, []string, error) {
	var R Info
	var basePath string
	var outFile []string //

	fi, err := os.Stat(path)
	if err != nil {
		return R, "", nil, err
	}

	basePath = filepath.ToSlash(filepath.Dir(path)) + `/` //文件
	ISDIR := false
	if fi.IsDir() {
		ISDIR = true
		path = filepath.ToSlash(path) + `/`
		basePath = filepath.ToSlash(filepath.Dir(filepath.Dir(path))) + `/` //文件夹
	}

	rmap := make(map[string]int64) //
	tmap := make(map[string]int64) //
	var tp string
	if ISDIR {
		err = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {

			if err != nil {
				if os.IsNotExist(err) {
					outFile = append(outFile, p)
					return nil
				} else if strings.Contains(err.Error(), `Access is denied.`) {
					outFile = append(outFile, p)
					return nil
				}
				return err
			}

			if info.IsDir() {
				return nil
			}
			if !info.Mode().IsRegular() {
				outFile = append(outFile, p)
				return nil
			}

			hl, err := os.Open(p)
			if err != nil {
				outFile = append(outFile, p)
				return nil
			}
			hl.Close()

			p, err = filepath.Rel(path, p)
			if err != nil {
				return err
			}
			tp = filepath.Base(path) + `/` + filepath.ToSlash(p)
			rmap[tp] = int64(info.Size())
			tmap[tp] = info.ModTime().UnixNano()
			return nil
		})
		if err != nil {
			return R, "", nil, err
		}
	} else {
		hl, err := os.Open(path)
		if err != nil {
			outFile = append(outFile, path)
			return R, "", nil, err
		}
		hl.Close()

		R.S = []int64{int64(fi.Size())}
		R.N = []string{filepath.Base(path)}
		R.T = []int64{int64(fi.ModTime().UnixNano())}
		return R, basePath, nil, nil
	}

	// sort
	var nameSlice []string
	for k := range rmap {
		nameSlice = append(nameSlice, k)
	}
	sort.Sort(sort.StringSlice(nameSlice))
	ls := len(nameSlice)

	sizeSlice := make([]int64, ls)
	timeSlice := make([]int64, ls)

	for i, v := range nameSlice {
		if i == 0 {
			sizeSlice[0] = rmap[v]
			timeSlice[0] = tmap[v]
		} else {
			sizeSlice[i] = rmap[v]
			timeSlice[i] = tmap[v]
		}
	}
	R.S = sizeSlice
	R.N = nameSlice
	R.T = timeSlice

	runtime.GC()

	return R, basePath, outFile, nil
}

// Wrap 各个系统下的换行符
func Wrap() string {
	if runtime.GOOS == "windows" {
		return "\r\n"
	} else if runtime.GOOS == "darwin" {
		return "\r"
	} else {
		return "\n"
	}
}
