// Things that should've been built-in to Go :) Intended to be imported with the "." operator.
// I feel dirty using "import .", so these primitives have lot to live up to..
package builtin

import (
	"fmt"
)

func FirstError(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}

	return nil
}

func UnsetErrorIf(isUnset bool, fieldName string) error {
	if isUnset {
		return fmt.Errorf("'%s' is required but not set", fieldName)
	} else {
		return nil
	}
}
// TODO: Coalesce() when we get generics..
