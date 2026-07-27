package repository

import "time"

// nowFn is a package-level seam so tests can override the clock.
// It defaults to time.Now in production code.
var nowFn = time.Now
