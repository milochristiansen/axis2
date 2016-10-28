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

package axis2_test

import (
	"fmt"
	"sort"
	"io/ioutil"
	"encoding/base64"
	
	"github.com/milochristiansen/axis2"
	"github.com/milochristiansen/axis2/sources/zip"
)

func Example() {
	// Create the filesystem
	fs := new(axis2.FileSystem)
	
	// Create our base data source.
	// There are several kinds of DataSource provided, here we will use the one that reads a zip file from
	// a byte slice. Most of the time you will want to use one that reads from an OS file or directory.
	ds, err := zip.NewRawDir(data)
	if err != nil {
		fmt.Println(err)
		return
	}
	// Then mount the data source as "base" for reading only.
	fs.Mount("base", ds, false)
	
	// At this point you would generally store "fs" somewhere you can reach it easily and use it over and over.
	// Keep in mind that you can mount many items on a single FileSystem, they don't even need unique names!
	// 
	// Items that do not have unique names will act like a single item with the caveat that each action will
	// be tried on them in the order they were mounted until it succeeds on one of them. For example you can
	// mount a user settings directory for reading and writing, then mount a directory with the default settings
	// for reading only. Any setting files the user has provided will be read from their directory and any
	// settings that the program writes will be written there, but anything that has not been provided will
	// be read from the default directory.
	// 
	// Anyway, back to our regularly scheduled example:
	
	// Now let's print the file tree... (you should read the code for this function!)
	printTree(fs, "", "")
	
	// ...and finally let's print the contents of "base/a/y.txt".
	// (note there are also ReadAll and WriteAll convenience methods, but I won't use
	// them here so you can see the whole process)
	rdr, err := fs.Read("base/a/y.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rdr.Close()
	contents, err := ioutil.ReadAll(rdr)
	fmt.Printf("\n%q\n", contents)
	
	// Errors from AXIS are wrapped in a special type, much like the "os" package warps most errors with the
	// os.PathError type.
	rdr, err = fs.Read("base/a/m.txt")
	if err == nil {
		rdr.Close()
		fmt.Println("Reading a non-existent file worked... Odd.")
		return
	}
	err2, ok := err.(*axis2.Error)
	if !ok {
		fmt.Println("AXIS returned an error of the incorrect type, this is supposed to be impossible.")
		return
	}
	
	// File not found errors from the "os" package are converted directly to the associated AXIS error type. Other
	// errors from os (that are wrapped in a os.PathError) are unwrapped and then rewrapped with the AXIS Error type.
	// This has the effect of stripping the OS path from the error message (the OS path is generally redundant and/or
	// undesirable, so this is intentional).
	if err2.Typ != axis2.ErrNotFound {
		fmt.Println("Error has unexpected type, this is supposed to be impossible.")
		return
	}
	fmt.Println("Errors OK!")
	
	// Output:
	// base/
	//   a/
	//     x.txt
	//     y.txt
	//     z.txt
	//   b.txt
	//   c.txt
	// 
	// "y.txt"
	// Errors OK!
}

func printTree(fs *axis2.FileSystem, p, d string) {
	// We need to sort the lists we get from ListDirs and ListFiles because the zip
	// file data sources are not guaranteed to list in the same order each time!
	// Normally a stable order is of no importance, but the test randomly failing due
	// to different output orders is not desirable...
	
	// List all the directories OR mount points in the current path.
	// Since we pass in the empty string as the initial path this should return a
	// list of the root mount point subsets the first time printTree is called
	// (in this case just "base"). Mount point subsets are generally treated like
	// directories, but it is an error to read or write something to them. Use IsMP
	// to see if a "directory" is actually a mount point subset (generally you will
	// already know unless you are blindly walking a tree like this).
	l := fs.ListDirs(d); sort.Strings(l)
	for _, dir  := range l {
		fmt.Println(p+dir+"/")
		printTree(fs, p+"  ", d+dir+"/")
	}
	
	// Rather than using ListDirs and ListFiles I could have just used List, but
	// then I would have had to use IsDir and IsMP to filter the results (and
	// directories wouldn't list first without some fancy code).
	l = fs.ListFiles(d); sort.Strings(l)
	for _, file  := range l {
		fmt.Println(p+file)
	}
}


// After init runs data will contain a zip file with the following contents:
//	a/x.txt
//	a/y.txt
//	a/z.txt
//	b.txt
//	c.txt
// Each file contains its name as an ASCII string.
var data []byte
func init() {
	// Yes, I am throwing the error away. How evil... (Don't try this at home!)
	data, _ = base64.StdEncoding.DecodeString(`
UEsDBBQAAAAAAHJgW0kAAAAAAAAAAAAAAAACAAAAYS9QSwMECgAAAAAACmFbSUkCG6wFAAAABQAAAAcA
AABhL3gudHh0eC50eHRQSwMECgAAAAAADWFbSfkre5EFAAAABQAAAAcAAABhL3kudHh0eS50eHRQSwME
CgAAAAAAEWFbSSlR29YFAAAABQAAAAcAAABhL3oudHh0ei50eHRQSwMECgAAAAAAFmFbSWqNS4YFAAAA
BQAAAAUAAABiLnR4dGIudHh0UEsDBAoAAAAAABlhW0napCu7BQAAAAUAAAAFAAAAYy50eHRjLnR4dFBL
AQI/ABQAAAAAAHJgW0kAAAAAAAAAAAAAAAACACQAAAAAAAAAEAAAAAAAAABhLwoAIAAAAAAAAQAYAPli
b6trMNIB+WJvq2sw0gFgGMOaazDSAVBLAQI/AAoAAAAAAAphW0lJAhusBQAAAAUAAAAHACQAAAAAAAAA
IAAAACAAAABhL3gudHh0CgAgAAAAAAABABgAxGEfVWww0gFrH32kazDSAWsffaRrMNIBUEsBAj8ACgAA
AAAADWFbSfkre5EFAAAABQAAAAcAJAAAAAAAAAAgAAAASgAAAGEveS50eHQKACAAAAAAAAEAGADkA0NZ
bDDSAa5cvqdrMNIBax99pGsw0gFQSwECPwAKAAAAAAARYVtJKVHb1gUAAAAFAAAABwAkAAAAAAAAACAA
AAB0AAAAYS96LnR4dAoAIAAAAAAAAQAYAOWx0l1sMNIBF0F2qmsw0gFrH32kazDSAVBLAQI/AAoAAAAA
ABZhW0lqjUuGBQAAAAUAAAAFACQAAAAAAAAAIAAAAJ4AAABiLnR4dAoAIAAAAAAAAQAYAGe6SWNsMNIB
2qQzmGsw0gHQ9qOVazDSAVBLAQI/AAoAAAAAABlhW0napCu7BQAAAAUAAAAFACQAAAAAAAAAIAAAAMYA
AABjLnR4dAoAIAAAAAAAAQAYANmuWGdsMNIB0PajlWsw0gHQ9qOVazDSAVBLBQYAAAAABgAGAA0CAADu
AAAAAAA=`)
}
