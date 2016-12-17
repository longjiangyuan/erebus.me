package control

import (
	"net/http"
	"strconv"
)

func FormInt(r *http.Request, name string, defVal int) int {
	s := r.FormValue(name)
	if s == "" {
		return defVal
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return defVal
	}
	return i
}
