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

type ErrTyp int
const (
	// The path does not point to a valid item.
	ErrNotFound ErrTyp = iota
	
	// The requested action could not be carried out because the item is read-only.
	ErrReadOnly
	
	// The action cannot be done with the item (for example trying to write a directory).
	ErrBadAction
	
	// The given path is not absolute.
	ErrPathRelative
	
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
	
	// TODO: Detect common OS errors and convert them directly to AXIS equivalents instead of wrapping.
	
	return &Error{
		Path: path,
		Typ: ErrRaw,
		Err: err,
	}
}

// Error wraps a path and an error type, together with an error from an external library if applicable.
// Errors returned by API functions will be of this type.
type Error struct {
	Path string // Automatically set by the API functions before return
	Typ ErrTyp
	
	// If Typ == ErrRaw
	Err error
}

func (err *Error) Error() string {
	switch err.Typ {
	case ErrNotFound:
		return "File/Dir Not Found: " + err.Path
	case ErrReadOnly:
		return "File/Dir Read-only: " + err.Path
	case ErrBadAction:
		return "The DataSource does not allow that action: " + err.Path
	case ErrPathRelative:
		return "The path is not absolute: " + err.Path
	case ErrRaw:
		return err.Err.Error() + " AXIS path: " + err.Path
	default:
		return "Invalid Error Code: " + err.Path
	}
}
