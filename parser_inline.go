package njson

func checkAnString(s string) bool {
	return len(s) > 2 && s[1] == 'a' && s[2] == 'N'
}

func checkUllString(data string) bool {
	return len(data) > 3 && data[1] == 'u' && data[2] == 'l' && data[3] == 'l'
}

func checkRueString(data string) bool {
	return len(data) > 3 && data[1] == 'r' && data[2] == 'u' && data[3] == 'e'
}

func checkAlseString(data string) bool {
	return len(data) > 4 && data[1] == 'a' && data[2] == 'l' && data[3] == 's' && data[4] == 'e'
}
