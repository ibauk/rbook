package main

import "strings"

func xmlsafe(s string) string {

	x := map[string]string{`&`: `&amp;`, `"`: `&quot;`, `<`: `&lt;`, `>`: `&gt;`, `'`: `&#39;`}
	res := s
	for k, v := range x {
		res = strings.ReplaceAll(res, k, v)
	}
	return res
}
