//go:build js

package store

import (
	"errors"
	"os"
)

// The browser build (the warranty verify page, ui-reviewer-spec §2.3) links this package through the
// connectors the warranty package imports (with its database driver: a 9 MB WebAssembly file, about
// 3 MB compressed; recorded, code review F6). The page never opens a store, so a lock always fails.
var errLockBusy = errors.New("store: no file locks in a browser")

func lockPlatform(*os.File) error { return errLockBusy }
