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

// AXIS VFS, a simple virtual file system API.
// 
// AXIS is based on a few simple interfaces and a set of API functions that operate on these interfaces.
// Clients use the provided implementations of these interfaces (or provide their own custom implementations)
// to create "data sources" that may be mounted on a "file system" and used for OS-independent file IO.
// 
// AXIS was originally written to allow files inside of archives to be handled with exactly the same API as
// used for files inside of directories, but it has since grown to allow "logical" files and directories as
// well as "multiplexing" multiple items on the same location (to, for example, make two directories look
// and act like one). These properties make AXIS perfect for handling data and configuration files for any
// program where flexibility is important, the program does not need to know where its files are actually
// located, it simply needs them to be at a certain place in it's AXIS file system. Changing where a program
// loads it's files from is then as simple as changing the code that initializes the file system.
// 
// AXIS uses standard slash separated paths. File names may not contain any "special" characters, basically
// any of the following:
// 
//	< > ? * | : " \ /
// 
// Additionally "." and ".." are not valid names.
// 
// Multiple slashes in an AXIS path are condensed into a single slash, and any trailing slashes will be stripped off.
// For example the following two paths are equivalent:
// 
//	test/path/to/dir
//	/test///path//to/dir/
// 
// Obviously you should always use the first form, but the second is still legal (barely)
// 
// AXIS VFS "officially" stands for Absurdly eXtremely Incredibly Simple Virtual File System (adjectives are
// good for making cool acronyms!). If you think the name is stupid (it is) you can just call it AXIS and
// forget what it is supposed to mean, after all the "official" name is more of a joke than anything...
package axis2

// Needed for a commented out debugging function
//import "strings"
//import "fmt"

import "io"
import "io/ioutil"

// DataSource is any item that implements either File or Dir (or, more rarely, both).
// 
// This property is enforced by the API, any functions that takes a DataSource will return an error if it does not
// implement the required interface(s), likewise functions that return a DataSource will always return a value that
// implements the required interface(s).
type DataSource interface {}

// Flags for "Dir.Child".
const (
	// Do not create the item if it does not exist.
	CreateNone int = iota
	
	// If the item does not exist try to create it as a child directory.
	CreateDir
	
	// If the item does not exist try to create it as a file.
	CreateFile
)

// Dir is the interface directories (or items that act like directories) must implement.
type Dir interface {
	// Child returns a reference to the requested child item, possibly creating it if needed and allowed.
	// 
	// If the item exists then return a reference to it (ignoring the value of create), else create the item using the
	// value of create as a hint.
	// 
	// Note that you do not need to actually create the item, simply creating the *possibility to have the item*
	// is enough. A request to create an item is always followed by a request to open it (or in the case of a
	// directory, one of its children) for writing, so you can delay the actual creation until then.
	Child(id string, create int) DataSource
	
	// Delete the given child item.
	Delete(id string) error
	
	// List all the children of this Dir.
	List() []string
}

// File is the interface files (or items that act like files) must implement.
type File interface {
	// Read opens an AXIS file for reading and returns the result and any error that may have happened.
	Read() (io.ReadCloser, error)
	
	// Write opens an AXIS file for writing and returns the result and any error that may have happened.
	// Any existing contents the file may have are truncated.
	Write() (io.WriteCloser, error)
	
	// Append is exactly like Write, except the file is not truncated before writing. 
	Append() (io.WriteCloser, error)
	
	// Size returns the size of the file or -1 if the value could not be retrieved.
	Size() int64
}

type source struct {
	mp []string
	ds DataSource
}

// FileSystem is the center of an AXIS setup.
// 
// FileSystems have two halves: A "read" half and a "write" half. Any action that changes something (Write, Delete, etc)
// is carried out on the "write" half and any action that involves reading existing information is carried out on the
// "read" half. For this reason most DataSources are mounted on both halves or just the read half, never on just the
// write half.
// 
// If you mount more than one item on a location they will be tried in order, the first one to work is the one that is
// used.
// 
// The zero value of FileSystem is an empty FileSystem ready to use.
type FileSystem struct {
	r []*source
	w []*source
}

/*
// Dump is a simple debugging function that list all resources mounted for reading to standard output.
func (fs *FileSystem) Dump() {
	for _, source := range fs.r {
		fmt.Printf("%q: (%T)%#v\n", strings.Join(source.mp, "/"), source.ds, source.ds)
	}
}
*/

// Mount a given DataSource onto the FileSystem at the given path.
// If rw is true the DataSource is mounted for writing as well as reading.
func (fs *FileSystem) Mount(path string, ds DataSource, rw bool) error {
	dirs := validatePath(path)
	if dirs == nil {
		return &Error{Path: path, Typ: ErrBadPath}
	}
	
	// Ensure the mounted item implements either File or Dir (or both).
	_, a := ds.(File); _, b := ds.(Dir)
	if !a && !b {
		return &Error{Path: path, Typ: ErrBadAction}
	}
	
	src := &source{
		mp: dirs,
		ds: ds,
	}
	fs.r = append(fs.r, src)
	if rw {
		fs.w = append(fs.w, src)
	}
	return nil
}

// Unmount deletes all mounted DataSources with the given mount point.
func (fs *FileSystem) Unmount(path string, r bool) error {
	dirs := validatePath(path)
	if dirs == nil {
		return &Error{Path: path, Typ: ErrBadPath}
	}
	
	fs.w = unmount(dirs, fs.w)
	if r {
		fs.r = unmount(dirs, fs.r)
	}
	return nil
}

func unmount(dirs []string, sources []*source) []*source {
	next:
	for i := 0; i < len(sources); {
		if len(dirs) != len(sources[i].mp) {
			i++
			continue
		}
		
		for k := range dirs {
			if dirs[k] != sources[i].mp[k] {
				i++
				continue next
			}
		}
		
		// Mount point matches the kill list, eliminate.
		copy(sources[i:], sources[i+1:])
		sources = sources[:len(sources)-1]
	}
	
	return sources
}

// SwapMount replaces the first data source with the given mount point and returns the old data source.
// Returns nil on error.
func (fs *FileSystem) SwapMount(path string, ds DataSource, rw bool) DataSource {
	dirs := validatePath(path)
	if dirs == nil {
		return nil
	}
	
	_, a := ds.(File); _, b := ds.(Dir)
	if !a && !b {
		return nil
	}
	
	rtn := remount(dirs, ds, fs.r)
	if rw {
		remount(dirs, ds, fs.w)
	}
	return rtn
}

func remount(dirs []string, ds DataSource, sources []*source) DataSource {
	next:
	for i := 0; i < len(sources); {
		if len(dirs) != len(sources[i].mp) {
			i++
			continue
		}
		
		for k := range dirs {
			if dirs[k] != sources[i].mp[k] {
				i++
				continue next
			}
		}
		
		// Mount point matches the list, replace.
		rtn := sources[i].ds
		sources[i].ds = ds
		return rtn
	}
	return nil
}

// Returns a list of mount point parts that begin with the given path.
// The names returned act like valid directory names for most purposes.
// Duplicates are elided.
func (fs *FileSystem) mountSubset(path string, r bool) []string {
	dirs := validatePath(path)
	if dirs == nil {
		return nil
	}
	
	var sources []*source
	if r {
		sources = fs.r
	} else {
		sources = fs.w
	}
	
	var rtn []string
	have := map[string]bool{}
	next:
	for _, src := range sources {
		if len(src.mp) <= len(dirs) {
			continue
		}
		
		for i := range dirs {
			if src.mp[i] != dirs[i] {
				continue next
			}
		}
		mp := src.mp[len(dirs)]
		if have[mp] {
			continue
		}
		have[mp] = true
		rtn = append(rtn, mp)
	}
	return rtn
}

// isMP does the same basic thing as mountSubset, but instead of making a list of mount
// points it simply returns true if there are any.
// 
// This is faster than mountSubset since it can return as soon as it finds something.
func (fs *FileSystem) isMP(path string, r bool) bool {
	dirs := validatePath(path)
	if dirs == nil {
		return false
	}
	
	var sources []*source
	if r {
		sources = fs.r
	} else {
		sources = fs.w
	}
	
	next:
	for _, src := range sources {
		if len(src.mp) <= len(dirs) {
			continue
		}
		
		for i := range dirs {
			if src.mp[i] != dirs[i] {
				continue next
			}
		}
		return true
	}
	return false
}

// GetDSAt returns the first DataSource that matches the given path.
// This is mostly for internal use, but it is useful for certain advanced actions.
// 
// If the path is a mount point subset with no DataSources mounted at that level then this
// will return an error with type ErrBadAction (ErrNotFound isn't appropriate in that case,
// because something exists at the path, just not a data source).
func (fs *FileSystem) GetDSAt(path string, create, r bool) (DataSource, error) {
	dss, err := fs.GetDSsAt(path, create, r)
	if err != nil {
		return nil, err
	}
	return dss[0], nil
}

// GetDSsAt returns all of the DataSources that match the given path.
// This is mostly for internal use, but it is useful for certain advanced actions.
// 
// If the path is a mount point subset with no DataSources mounted at that level then this
// will return an error with type ErrBadAction (ErrNotFound isn't appropriate in that case,
// because something exists at the path, just not a data source).
func (fs *FileSystem) GetDSsAt(path string, create, r bool) ([]DataSource, error) {
	dirs := validatePath(path)
	if dirs == nil {
		return nil, &Error{Path: path, Typ: ErrBadPath}
	}
	
	var sources []*source
	if r {
		sources = fs.r
	} else {
		sources = fs.w
	}
	
	var dss []DataSource
	
	next:
	for _, src := range sources {
		// First make sure that the path is a superset of the current source's mount point.
		i := 0
		for ; i < len(src.mp); i++ {
			if i >= len(dirs) || src.mp[i] != dirs[i] {
				continue next
			}
		}
		
		// Then try to get a child item from the source that matches the remainder of the path.
		ds := src.ds
		var pds DataSource
		for ; i < len(dirs); i++ {
			pds = ds
			pdir, ok := pds.(Dir)
			if !ok {
				continue next
			}
			
			c := CreateNone
			if create && i == len(dirs)-1 {
				c = CreateFile
			} else if create {
				c = CreateDir
			}
			
			ds = pdir.Child(dirs[i], c)
			if ds == nil {
				continue next
			}
		}
		
		dss = append(dss, ds)
	}
	
	if dss != nil {
		return dss, nil
	}
	if fs.isMP(path, r) {
		return nil, &Error{Path: path, Typ: ErrBadAction}
	}
	return nil, &Error{Path: path, Typ: ErrNotFound}
}

// Exists returns true if the path points to a valid DataSource or a mount point subset.
// If the queried item is a mount point subset you won't be able to read or write it!
func (fs *FileSystem) Exists(path string) bool {
	_, err := fs.GetDSAt(path, false, true)
	if err != nil {
		// Report that the path exists (even if it doesn't really) when it is a subset of one or more mount points.
		return fs.isMP(path, true)
	}
	return true
}

// IsDir returns true if the path points to a valid Dir or mount point subset.
func (fs *FileSystem) IsDir(path string) bool {
	ds, err := fs.GetDSAt(path, false, true)
	if err != nil {
		// Report that the path is a directory (even if it isn't really) when it is a subset of one or more mount points.
		return fs.isMP(path, true)
	}
	
	_, ok := ds.(Dir)
	return ok
}

// IsMP returns true if the path is a mount point or mount point subset.
func (fs *FileSystem) IsMP(path string) bool {
	// The external version does not let you specify which half to check, instead it always checks the read
	// half (the write half is always a subset of the read half anyway).
	return fs.isMP(path, true)
}

// Size returns the size of the File at the given path. If the object is not a File or the path is invalid -1 is returned.
func (fs *FileSystem) Size(path string) int64 {
	ds, err := fs.GetDSAt(path, false, true)
	if err != nil {
		return -1
	}
	
	f, ok := ds.(File)
	if !ok {
		return -1
	}
	return f.Size()
}

// Delete attempts to delete the item at the given path. This may or may not work. Deleting is always carried out on the
// write portion of the FileSystem, objects on the read portion will not be effected unless they are also mounted for
// writing. Only the first item found is deleted.
func (fs *FileSystem) Delete(path string) error {
	npath, last := trimLastPath(path)
	
	dss, err := fs.GetDSsAt(npath, false, false)
	if err != nil {
		return err
	}
	
	for _, ds := range dss {
		d, ok := ds.(Dir)
		if ok {
			if d.Child(last, 0) != nil {
				return wrapError(d.Delete(last), path)
			}
		}
	}
	return &Error{Path: path, Typ: ErrNotFound}
}

// List returns a slice of the names of all the items available in the given Dir. If the item at the path is not a Dir
// and is not a subset of any mount points this returns nil.
// 
// The order of the returned list is undefined, or more correctly, is defined by the individual Dir implementations.
// Most of the time this means lexically by filename, but not always.
// 
// If the path is a mount point subset this may return more mount point subsets or a mix of mount point subsets and
// data sources!
func (fs *FileSystem) List(path string) []string {
	dss, err := fs.GetDSsAt(path, false, true)
	if err != nil {
		// Treat the path like a directory of directories if it is a mount point subset.
		return fs.mountSubset(path, true)
	}
	
	have := map[string]bool{}
	var rtn []string
	for _, ds := range dss {
		d, ok := ds.(Dir)
		if ok {
			for _, item := range d.List() {
				if !have[item] {
					have[item] = true
					rtn = append(rtn, item)
				}
			}
		}
	}
	// This should only matter in cases where mount points overlap data sources.
	// Appending a nil slice to a nil slice results in a nil slice, yes I checked.
	return append(rtn, fs.mountSubset(path, true)...)
}

// ListDirs returns a slice of all the items that are Dirs in the given Dir. If the item at the path is not a Dir
// or mount point subset this returns nil.
// 
// The order of the returned list is undefined, or more correctly, is defined by the individual Dir implementations.
// Most of the time this means lexically by filename, but not always.
// 
// If the path is a mount point subset this may return more mount point subsets or a mix of mount point subsets and
// data sources!
func (fs *FileSystem) ListDirs(path string) []string {
	dss, err := fs.GetDSsAt(path, false, true)
	if err != nil {
		// Treat the path like a directory of directories if it is a mount point subset.
		return fs.mountSubset(path, true)
	}
	
	have := map[string]bool{}
	var rtn []string
	for _, ds := range dss {
		d, ok := ds.(Dir)
		if !ok {
			continue
		}
		children := d.List()
		
		for _, child := range children {
			if !have[child] {
				have[child] = true
				cds := d.Child(child, 0)
				if _, ok := cds.(Dir); ok {
					rtn = append(rtn, child)
				}
			}
		}
	}
	// This should only matter in cases where mount points overlap data sources.
	// Appending a nil slice to a nil slice results in a nil slice, yes I checked.
	return append(rtn, fs.mountSubset(path, true)...)
}

// ListFiles returns a slice of all the items that are Files in the given Dir. If the item at the path is not a Dir
// or if the Dir contains no Files this returns nil.
// 
// The order of the returned list is undefined, or more correctly, is defined by the individual Dir implementations.
// Most of the time this means lexically by filename, but not always.
func (fs *FileSystem) ListFiles(path string) []string {
	dss, err := fs.GetDSsAt(path, false, true)
	if err != nil {
		return nil
	}
	
	have := map[string]bool{}
	var rtn []string
	for _, ds := range dss {
		d, ok := ds.(Dir)
		if !ok {
			return nil
		}
		children := d.List()
		
		for _, child := range children {
			if !have[child] {
				have[child] = true
				cds := d.Child(child, 0)
				if _, ok := cds.(File); ok {
					rtn = append(rtn, child)
				}
			}
		}
	}
	return rtn
}

// Read opens the File at the given path for reading.
func (fs *FileSystem) Read(path string) (io.ReadCloser, error) {
	ds, err := fs.GetDSAt(path, false, true)
	if err != nil {
		return nil, err
	}
	
	f, ok := ds.(File)
	if !ok {
		return nil, &Error{Path: path, Typ: ErrBadAction}
	}
	
	rc, err := f.Read()
	return rc, wrapError(err, path)
}

// ReadAll reads the File at the given path and returns it's contents.
func (fs *FileSystem) ReadAll(path string) ([]byte, error) {
	reader, err := fs.Read(path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	
	content, err := ioutil.ReadAll(reader)
	return content, wrapError(err, path)
}

// Write opens the the File at the given path for writing. Any existing file contents are truncated.
func (fs *FileSystem) Write(path string) (io.WriteCloser, error) {
	ds, err := fs.GetDSAt(path, true, false)
	if err != nil {
		return nil, err
	}
	
	f, ok := ds.(File)
	if !ok {
		return nil, &Error{Path: path, Typ: ErrBadAction}
	}
	
	wc, err := f.Write()
	return wc, wrapError(err, path)
}

// Append opens the file at the given path for writing. The write cursor is set beyond any existing file contents.
func (fs *FileSystem) Append(path string) (io.WriteCloser, error) {
	ds, err := fs.GetDSAt(path, true, false)
	if err != nil {
		return nil, err
	}
	
	f, ok := ds.(File)
	if !ok {
		return nil, &Error{Path: path, Typ: ErrBadAction}
	}
	
	wc, err := f.Append()
	return wc, wrapError(err, path)
}

// WriteAll replace the contents of the File at the given path with the contents given.
func (fs *FileSystem) WriteAll(path string, content []byte) error {
	writer, err := fs.Write(path)
	if err != nil {
		return err
	}
	defer writer.Close()
	
	_, err = writer.Write(content)
	return wrapError(err, path)
}
