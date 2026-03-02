package httpx

import (
	"net/http"
	"strings"
)

func HeaderToMap(h http.Header) map[string]string {
	m := make(map[string]string)
	for k, v := range h {
		m[k] = strings.Join(v, ", ")
	}
	return m
}
