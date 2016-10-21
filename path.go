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

package axis2

import "strings"

func trimLastPath(path string) (string, string) {
	i := strings.LastIndex(path, "/")
	if i == -1 {
		return "", path
	}
	return path[:i], path[i+1:]
}

func validatePath(path string) []string {
	if path == "" {
		return []string{}
	}
	
	if strings.ContainsAny(path, "<>?*|:\"\\") {
		return nil
	}
	
	dirs := strings.Split(path, "/")
	
	for _, dir := range dirs {
		if dir == ".." || dir == "." {
			return nil
		}
	}
	
	// Remove empty dir names
	for i := 0; i < len(dirs); {
		if dirs[i] != "" {
			i++
			continue
		}
		
		copy(dirs[i:], dirs[i+1:])
		dirs = dirs[:len(dirs)-1]
	}
	
	return dirs
}
