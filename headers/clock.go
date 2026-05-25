package headers

import "time"

// nowUnix is a seam so tests can pin the clock when needed.
var nowUnix = func() int64 {
	return time.Now().Unix()
}
