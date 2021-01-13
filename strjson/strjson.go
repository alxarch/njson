package strjson

import "strings"

type Flags uint8

const (
	FlagValid Flags = 1 << iota
	FlagSafe
	FlagJSON
	FlagHTML
)

func (f Flags) IsValid() bool {
	return f&(FlagValid) != 0
}
func (f Flags) IsJSON() bool {
	return f&(FlagJSON) == FlagJSON
}
func (f Flags) IsJSONSafe() bool {
	return f&(FlagSafe|FlagJSON) != 0
}

func (f Flags) IsGoSafe() bool {
	return f&(FlagSafe|FlagJSON) != FlagJSON
}

func (f Flags) IsHTML() bool {
	return f&FlagHTML == FlagHTML
}

type String struct {
	Value string
	flags Flags
}

func FromParts(s string, flags Flags) String {
	return String{
		Value: s,
		flags: flags,
	}
}

func FromJSON(s string, safe bool) String {
	if safe {
		return String{
			Value: s,
			flags: FlagJSON | FlagValid | FlagSafe,
		}
	}
	return String{
		Value: s,
		flags: FlagJSON | FlagValid,
	}
}

func FromSafeString(s string) String {
	return String{
		Value: s,
		flags: FlagSafe | FlagValid,
	}
}

func FromString(s string) String {
	return String{
		Value: s,
		flags: FlagValid,
	}
}
func FromHTML(s string) String {
	return String{
		Value: s,
		flags: FlagHTML | FlagValid,
	}
}

func (s String) String() string {
	if s.flags.IsGoSafe() {
		return s.Value
	}
	return Unescaped(s.Value)
}

func (s String) Flags() Flags {
	return s.flags
}

func (s String) IsValid() bool {
	return s.flags.IsValid()
}

func (s String) Unescape() String {
	if s.flags.IsGoSafe() {
		return s
	}
	return String{
		Value: Unescaped(s.Value),
		flags: s.flags &^ FlagJSON,
	}
}

func (s String) Escape(html bool) String {
	if s.flags.IsJSONSafe() {
		return s
	}
	e := Escaped(s.Value, html || s.flags.IsHTML(), false)
	esc := FromJSON(e, strings.IndexByte(e, '\\') == -1)
	esc.flags |= s.flags
	return esc
}
