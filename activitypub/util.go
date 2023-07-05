package activitypub

import "strings"

func parseResourceString(s string) (name string, resType string, url string) {
	var delim string
	if strings.Contains(s, "://") {
		delim = "://"
	} else if strings.Contains(s, ":") {
		delim = ":"
	} else {
		return
	}
	splt := strings.Split(s, delim)
	if len(splt) != 2 {
		return
	}
	resType = splt[0]
	name = splt[1]
	name, _ = strings.CutPrefix(name, "@")
	splt = strings.Split(name, "@")
	if len(splt) != 2 {
		return
	}
	name = splt[0]
	url = splt[1]
	return
}
