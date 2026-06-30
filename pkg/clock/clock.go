// Package clock menyediakan "virtual now" yang bisa dimajukan untuk simulasi waktu
// (kebutuhan overdue handling di Level 6). Semua logika bisnis yang butuh waktu
// (expiry diskon, deadline pengiriman) WAJIB pakai clock.Now(), bukan time.Now().
package clock

import (
	"sync/atomic"
	"time"
)

// offsetNanos adalah selisih (ns) antara virtual now dan waktu nyata.
var offsetNanos atomic.Int64

// Now mengembalikan waktu virtual = waktu nyata + offset.
func Now() time.Time {
	return time.Now().Add(time.Duration(offsetNanos.Load()))
}

// Offset mengembalikan offset saat ini.
func Offset() time.Duration {
	return time.Duration(offsetNanos.Load())
}

// SetOffset mengeset offset absolut (dipakai saat load dari DB).
func SetOffset(d time.Duration) {
	offsetNanos.Store(int64(d))
}

// Advance menambah offset (mis. maju 1 hari) dan mengembalikan offset baru.
func Advance(d time.Duration) time.Duration {
	return time.Duration(offsetNanos.Add(int64(d)))
}
