package base62

import "slices"

const base62Charset = "jV0XyWg9s8cQ5q2e4Z1tIuB7p3DkRzLnfF6hYvTmaNwbMOrAoHlUxJdCPKGiES"

func Base62Encode(id uint64) string {
	var code []byte
	if id == 0 {
		code = append(code, '0')
		return string(code)
	}
	for id > 0 {
		code = append(code, base62Charset[id%62])
		id /= 62
	}
	slices.Reverse(code)
	return string(code)
}
