package orderutils

import (
	crand "crypto/rand"
	"encoding/binary"
	"strconv"
	"time"
)

// GenerateOrderSalt returns a random positive 63-bit integer rendered as a
// decimal string, suitable as the EIP-712 uint256 salt field.
//
// Uses crypto/rand to avoid collisions when many orders are produced inside
// the same millisecond. Falls back to the nanosecond clock if the RNG fails
// (extremely rare in practice).
func GenerateOrderSalt() string {
	var b [8]byte
	if _, err := crand.Read(b[:]); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	// Drop the top bit so the value fits in int64.
	return strconv.FormatUint(binary.BigEndian.Uint64(b[:])>>1, 10)
}
