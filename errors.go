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

import "os"

type ErrTyp int
const (
	// The path does not point to a valid item.
	ErrNotFound ErrTyp = iota
	
	// The requested action could not be carried out because the item is read-only.
	ErrReadOnly
	
	// The action cannot be done with the item (for example trying to write a directory).
	ErrBadAction
	
	// The given path is not absolute or it contains invalid characters.
	ErrBadPath
	
	// An error from an external library, with an attached AXIS path.
	ErrRaw
)

// NewError creates a new AXIS Error with the given type.
// Error path information is automatically filled in by the API just before it is returned to the user.
func NewError(typ ErrTyp) error {
	return &Error{
		Typ: typ,
	}
}

// wrapError wraps an error from an external library and/or sets the Path
// field as required.
func wrapError(err error, path string) error {
	if err == nil {
		return nil
	}
	if e, ok := err.(*Error); ok {
		e.Path = path
		return e
	}
	
	// Don't wrap os.PathError values directly (we don't want the error message to include the OS path).
	if e, ok := err.(*os.PathError); ok {
		// Convert file not found directly to the equivalent AXIS error
		if e.Err == os.ErrNotExist {
			return &Error{
				Path: path,
				Typ: ErrNotFound,
			}
		}
		
		// Unwrap, then rewrap any other PathErrors
		return &Error{
			Path: path,
			Typ: ErrRaw,
			Err: e.Err,
		}
	}
	
	// Unknown error, just wrap it.
	return &Error{
		Path: path,
		Typ: ErrRaw,
		Err: err,
	}
}

// Error wraps a path and an error type, together with an error from an external library if applicable.
// Errors returned by API functions will be of this type.
// 
// Implementers of File or Dir do not need to use this type for their return values, but they are encouraged
// to do so (any returned error will be wrapped if it is not already of this type).
type Error struct {
	// Automatically set by the API functions before return, don't write this field!
	Path string
	
	// The error type. If you want to intelligently handle errors from this library this field will be invaluable.
	// See the ErrTyp constants for a list of valid values for this field.
	Typ ErrTyp
	
	// If Typ == ErrRaw this will contain an error value originating from a specific implementation of File or Dir.
	Err error
}

// Error prints the path associated with the error prefixed by a short string explanation of the error type.
// 
// Raw (wrapped) errors simply have "AXIS path: <path>" tacked onto the output of their own Error function.
func (err *Error) Error() string {
	switch err.Typ {
	case ErrNotFound:
		return "No item found at path: " + err.Path
	case ErrReadOnly:
		return "Item at path is read-only: " + err.Path
	case ErrBadAction:
		return "Illegal action for item at path: " + err.Path
	case ErrBadPath:
		return "Path is invalid: " + err.Path
	case ErrRaw:
		return err.Err.Error() + " AXIS path: " + err.Path
	default:
		return "Invalid error code: " + err.Path
	}
}
