package base62

import "slices"

const base62Charset = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

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
