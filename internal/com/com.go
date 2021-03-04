package com

import (
	"io"
	"log"
	"os"
	"runtime"
	"strconv"
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
