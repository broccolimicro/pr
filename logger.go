package main

import (
	"io"
	"os"
	"fmt"
	"strconv"
	"reflect"
	"path/filepath"
)

type ChannelLogger interface {
	Write(value interface{})
	io.Closer
}

type logger struct {
	dir string
	name string
	files map[string]*os.File
}

func Log(dir string, name string) ChannelLogger {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		fmt.Println(err)
	}

	return &logger {
		dir: dir,
		name: name,
		files: make(map[string]*os.File),
	}
}

func LogArr(n int, dir string, name string) []ChannelLogger {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		fmt.Println(err)
	}

	result := make([]ChannelLogger, n)
	for i := 0; i < n; i++ {
		result[i] = &logger {
			dir: dir,
			name: name + "." + strconv.Itoa(i),
			files: make(map[string]*os.File),
		}
	}
	return result
}

func (log *logger) Write(value interface{}) {
	log.WriteElem(log.name, value)
}

func (log *logger) WriteElem(caller string, value interface{}) {
	t := reflect.TypeOf(value)
	v := reflect.ValueOf(value)
	if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		for i := 0; i < v.Len(); i++ {
			log.WriteElem(caller + "." + strconv.Itoa(i), v.Index(i).Interface())
		}
	} else if t.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			log.WriteElem(caller + "." + t.Field(i).Name, v.Field(i).Interface())
		}
	} else {
		fptr, ok := log.files[caller]
		if !ok {
			var err error
			fptr, err = os.Create(filepath.Join(log.dir, caller+".dat"))
			if err != nil {
				fmt.Println(err)
				return
			}
			log.files[caller] = fptr
		}

		if b, ok := value.(bool); ok {
			if b {
				fmt.Fprintln(fptr, 1)
			} else {
				fmt.Fprintln(fptr, 0)
			}
		} else {
			fmt.Fprintln(fptr, value)
		}
	}
}

func (log *logger) Close() error {
	for key, fptr := range log.files {
		if fptr != nil {
			err := fptr.Close()
			if err != nil {
				return err
			}
			delete(log.files, key)
		}
	}
	return nil
}
