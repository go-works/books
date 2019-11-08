package main

// for a given file, get output of executing this command
// We cache this as it is the most expensive part of rebuilding books
// If allowError is true, we silence an error from executed command
// This is useful when e.g. executing "go run" on a program that is
// intentionally not valid.
func getOutputCached(cache *Cache, sf *SourceFile) error {
	// TODO: make the check for when not to execute the file
	// even better
	if sf.Directive.NoOutput {
		if sf.Directive.Glot {
			// if running on glot, we want to execute even if
			// we don't show the output (to check syntax errors)
		} else {
			return nil
		}
	}
	panic("shouldn't happen anymore")
	return nil
}
