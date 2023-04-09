package repo

import "regexp"

func IsTS(s string) bool {
	r := regexp.MustCompile(`^[0-2]\d.[0-2]\d.[0-2]\d\d\d [0-2]\d_\d\d_\d\d.\d{3,4}`)
	//logger.L.Infof("in store.IsTS %s matched? %t\n", s, res)
	return r.MatchString(s)
}
