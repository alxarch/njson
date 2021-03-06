package njson

import (
	"testing"
)

func TestIsDigit(t *testing.T) {
	for i := 0; i < 255; i++ {
		want := '0' <= i && i <= '9'
		if isDigit(byte(i)) != want {
			t.Fatalf("IsDigit(%c) != %t", i, want)
		}
	}
}

func TestIsNumberEnd(t *testing.T) {
	type args struct {
		c byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{",", args{','}, true},
		{" ", args{' '}, true},
		{"\n", args{'\n'}, true},
		{"\r", args{'\r'}, true},
		{"\t", args{'\t'}, true},
		{"}", args{'}'}, true},
		{"]", args{']'}, true},
		{"a", args{'a'}, false},
		{"0", args{'0'}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNumberEnd(tt.args.c); got != tt.want {
				t.Errorf("IsNumberEnd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsSpaceASCII(t *testing.T) {
	for i := 0; i < 255; i++ {
		want := false
		switch i {
		case '\r', '\n', ' ', '\t':
			want = true
		}
		if isSpace(byte(i)) != want {
			t.Fatalf("isSpave(%c) != %t", i, want)
		}
	}
}
