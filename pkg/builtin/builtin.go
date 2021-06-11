// Things that should've been built-in to Go :) Intended to be imported with the "." operator.
// I feel dirty using "import .", so these primitives have lot to live up to..
package builtin

import (
	"fmt"
)

// a type that cannot contain no information whatsoever.
// perfect for when you need to make(chan struct{}) purely for signalling etc.
type Void struct{}

func FirstError(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}

	return nil
}

// TODO: this would greatly benefit from generics
func FirstNonEmpty(items ...string) string {
	for _, item := range items {
		if item != "" {
			return item
		}
	}

	return ""
}

func UnsetErrorIf(isUnset bool, fieldName string) error {
	if isUnset {
		return fmt.Errorf("'%s' is required but not set", fieldName)
	} else {
		return nil
	}
}
