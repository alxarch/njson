package numjson

import (
	"math"
	"strconv"
)

// AppendFloat float appends the string of a float to a byte slice following rules commonly used by JSON formatters.
func AppendFloat(dst []byte, f float64, bits int) []byte {
	abs := math.Abs(f)
	fmt := byte('f')
	if abs != 0 && ((bits == 32 && (float32(abs) < 1e-6 || float32(abs) >= 1e21)) ||
		(bits == 64 && (abs < 1e-6 || abs >= 1e21))) {
		fmt = 'e'
	}
	dst = strconv.AppendFloat(dst, f, fmt, -1, bits)
	if fmt == 'e' {
		if i := len(dst) - 4; 0 <= i && i < len(dst) {
			if buf := dst[i:]; len(buf) == 4 && buf[0] == 'e' && buf[1] == '-' && buf[2] == '0' {
				buf[2] = buf[3]
				if i += 3; 0 <= i && i < len(dst) {
					dst = dst[:i]
				}
			}
		}
	}
	return dst
}

// FormatFloat returns the string of a float following rules commonly used by JSON formatters.
func FormatFloat(f float64, bits int) string {
	b := make([]byte, 0, 64)
	b = AppendFloat(b, f, bits)
	return string(b)
}
