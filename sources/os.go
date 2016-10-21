/*
Copyright 2016 by Milo Christiansen

This software is provided 'as-is', without any express or implied warranty. In
no event will the authors be held liable for any damages arising from the use of
this software.

Permission is granted to anyone to use this software for any purpose, including
commercial applications, and to alter it and redistribute it freely, subject to
the following restrictions:

1. The origin of this software must not be misrepresented; you must not claim
that you wrote the original software. If you use this software in a product, an
acknowledgment in the product documentation would be appreciated but is not
required.

2. Altered source versions must be plainly marked as such, and must not be
misrepresented as being the original software.

3. This notice may not be removed or altered from any source distribution.
*/

package sources

import ospath "path"
import "os"
import "io"
import "io/ioutil"

import "dctech/axis2"

type osFile string

// NewOSFile creates a new OS AXIS file interface.
func NewOSFile(path string) axis2.File {
	return osFile(path)
}

func (file osFile) Size() int64 {
	path := string(file)
	
	s, err := os.Lstat(path)
	if err != nil {
		return -1
	}
	return s.Size()
}

func (file osFile) Read() (io.ReadCloser, error) {
	path := string(file)
	
	return os.Open(path)
}

func (file osFile) Write() (io.WriteCloser, error) {
	path := string(file)
	
	_, err := os.Stat(path)
	if err != nil {
		err := os.MkdirAll(ospath.Dir(path), 0777)
		if err != nil {
			return nil, err
		}
		return os.Create(path)
	}
	
	return os.Create(path)
}

func (file osFile) Append() (io.WriteCloser, error) {
	path := string(file)
	
	_, err := os.Stat(path)
	if err != nil {
		err := os.MkdirAll(ospath.Dir(path), 0777)
		if err != nil {
			return nil, err
		}
		return os.Create(path)
	}
	
	return os.OpenFile(path, os.O_WRONLY, 0)
}

type osDir string

// NewOSDir creates a new OS AXIS directory interface.
func NewOSDir(path string) axis2.Dir {
	return osDir(path)
}

func (dir osDir) Child(id string, create int) axis2.DataSource {
	path := string(dir) + "/" + id
	
	info, err := os.Stat(path)
	if err == nil {
		if info.IsDir() {
			return osDir(path)
		}
		return osFile(path)
	}
	switch create {
	case axis2.CreateDir:
		return osDir(path)
	case axis2.CreateFile:
		return osFile(path)
	default:
		return nil
	}
}

func (dir osDir) Delete(id string) error {
	path := string(dir)
	
	return os.Remove(path + "/" + id)
}

func (dir osDir) List() []string {
	path := string(dir)
	
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil
	}
	
	rtn := make([]string, 0, len(files))
	for _, file := range files {
		rtn = append(rtn, file.Name())
	}
	return rtn
}

