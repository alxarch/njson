package njson_test

import (
	"testing"
	"unicode"

	"github.com/alxarch/njson"
)

func TestToHexDigit(t *testing.T) {
	d := njson.ToHexDigit('d')
	if d != 13 {
		t.Errorf("Invalid hex digit: %d", d)
	}
}

func TestIsDigit(t *testing.T) {
	for i := 0; i < 255; i++ {
		want := '0' <= i && i <= '9'
		t.Run(string([]byte{byte(i)}), func(t *testing.T) {
			if got := njson.IsDigit(byte(i)); got != want {
				t.Errorf("IsDigit() = %v, want %v", got, want)
			}
		})
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
			if got := njson.IsNumberEnd(tt.args.c); got != tt.want {
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
		t.Run(string([]byte{byte(i)}), func(t *testing.T) {
			if got := njson.IsSpaceASCII(byte(i)); got != want {
				t.Errorf("IsSpaceASCII() = %v, want %v", got, want)
			}
		})
	}
}

func TestToHex(t *testing.T) {
	type args struct {
		c byte
	}
	tests := []struct {
		name string
		args args
		want byte
	}{
		{"0", args{0}, '0'},
		{"1", args{1}, '1'},
		{"2", args{2}, '2'},
		{"3", args{3}, '3'},
		{"4", args{4}, '4'},
		{"5", args{5}, '5'},
		{"6", args{6}, '6'},
		{"7", args{7}, '7'},
		{"8", args{8}, '8'},
		{"9", args{9}, '9'},
		{"10", args{10}, 'a'},
		{"11", args{11}, 'b'},
		{"12", args{12}, 'c'},
		{"13", args{13}, 'd'},
		{"14", args{14}, 'e'},
		{"15", args{15}, 'f'},
		{"16", args{16}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := njson.ToHex(tt.args.c); got != tt.want {
				t.Errorf("ToHex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToLowerASCII(t *testing.T) {
	for i := 0; i < 255; i++ {
		c := byte(i)
		want := byte(unicode.ToLower(rune(c)))
		t.Run(string(c), func(t *testing.T) {
			if got := njson.ToLowerASCII(c); got != want {
				t.Errorf("ToLowerASCII() = %v, want %v", got, want)
			}
		})

	}
}

func TestToNamedEscape(t *testing.T) {
	type args struct {
		c byte
	}
	tests := []struct {
		name string
		args args
		want byte
	}{
		{"\\n", args{'n'}, '\n'},
		{"\\r", args{'r'}, '\r'},
		{"\\t", args{'t'}, '\t'},
		{"\\b", args{'b'}, '\b'},
		{"\\f", args{'f'}, '\f'},
		{"a", args{'a'}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := njson.ToNamedEscape(tt.args.c); got != tt.want {
				t.Errorf("ToNamedEscape() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToUpperASCII(t *testing.T) {
	for i := 0; i < 255; i++ {
		c := byte(i)
		want := byte(unicode.ToUpper(rune(c)))
		t.Run(string(c), func(t *testing.T) {
			if got := njson.ToUpperASCII(c); got != want {
				t.Errorf("ToLowerASCII() = %v, want %v", got, want)
			}
		})

	}
}
