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

package zip

import "dctech/axis2"

import "io"
import "bytes"
import "strings"
import "archive/zip"

type zdir struct {
	items map[string]interface{} // Either *zdir or *zfile
	zip   *zip.Reader
}

type zfile struct {
	me  *zip.File
	zip *zip.Reader
}

// NewDir creates a read-only AXIS Dir backed by a zip file.
func NewDir(file io.ReaderAt, size int64) (axis2.Dir, error) {
	z, err := zip.NewReader(file, size)
	if err != nil {
		return nil, err
	}
	return mkTree(z), nil
}

// NewRawDir creates a read-only AXIS Dir backed by a zip file that has been read into memory.
func NewRawDir(content []byte) (axis2.Dir, error) {
	file := bytes.NewReader(content)
	
	z, err := zip.NewReader(file, int64(file.Len()))
	if err != nil {
		return nil, err
	}
	return mkTree(z), nil
}

// Since zip files are assumed readonly I generate a static tree of dir and file objects when opening the zip.
// This makes file lookup much faster.
func mkTree(z *zip.Reader) *zdir {
	base := &zdir{
		items: map[string]interface{}{},
		zip: z,
	}
	
	for _, file := range z.File {
		parts := strings.Split(strings.TrimRight(file.Name, "/"), "/")
		dir := base
		for i := 0; i < len(parts)-1; i++ {
			child, ok := dir.items[parts[i]]
			if !ok {
				// This will almost certainly never trigger.
				child = &zdir{
					items: map[string]interface{}{},
					zip: z,
				}
				dir.items[parts[i]] = child
			}
			dir = child.(*zdir)
		}
		if file.FileInfo().IsDir() {
			dir.items[parts[len(parts)-1]] = &zdir{
				items: map[string]interface{}{},
				zip: z,
			}
		} else {
			dir.items[parts[len(parts)-1]] = &zfile{
				me: file,
				zip: z,
			}
		}
	}
	return base
}

func (dir *zdir) Child(id string, create int) axis2.DataSource {
	return dir.items[id]
}

func (dir *zdir) Delete(id string) error {
	return axis2.NewError(axis2.ErrReadOnly)
}

func (dir *zdir) List() []string {
	var rtn []string
	for n := range dir.items {
		rtn = append(rtn, n)
	}
	return rtn
}

func (file *zfile) Size() int64 {
	return int64(file.me.UncompressedSize64)
}

func (file *zfile) Read() (io.ReadCloser, error) {
	return file.me.Open()
}

func (file *zfile) Write() (io.WriteCloser, error) {
	return nil, axis2.NewError(axis2.ErrReadOnly)
}

func (file *zfile) Append() (io.WriteCloser, error) {
	return nil, axis2.NewError(axis2.ErrReadOnly)
}
