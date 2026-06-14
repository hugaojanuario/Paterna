package errorsx

import "errors"

var (
	ErrNotFound             = errors.New("not found")
	ErrContainersNotRunning = errors.New("containers not running")
)
