package strjson

import (
	"testing"
)

func TestToHex(t *testing.T) {
	for b, x := range map[byte]byte{
		0:  '0',
		1:  '1',
		2:  '2',
		3:  '3',
		4:  '4',
		5:  '5',
		6:  '6',
		7:  '7',
		8:  '8',
		9:  '9',
		10: 'A',
		11: 'B',
		12: 'C',
		13: 'D',
		14: 'E',
		15: 'F',
	} {
		if toHex(b) != x {
			t.Errorf("Invalid hex byte %c != %c", b, x)
		}
	}
}
