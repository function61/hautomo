// Things that should've been built-in to Go :) Intended to be imported with the "." operator.
// I feel dirty using "import .", so these primitives have lot to live up to..
package builtin

func FirstError(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}

	return nil
}

// TODO: Coalesce() when we get generics..
