package identity

import "time"

// nowUTC is a single seam for "current time" so it's the one place a
// future test would need to fake the clock.
func nowUTC() time.Time { return time.Now().UTC() }
