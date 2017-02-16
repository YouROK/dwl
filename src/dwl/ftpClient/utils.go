package ftpClient

import (
	"errors"
	"net/url"
	"strings"
)

func parseUrl(rawurl string) (host string, path string, user string, pass string, err error) {
	var u *url.URL
	u, err = url.Parse(rawurl)
	if err != nil {
		return "", "", "", "", err
	}
	if u.Scheme != "ftp" {
		return "", "", "", "", errors.New("not ftp protocol scheme")
	}

	host = u.Host
	if len(strings.Split(host, ":")) == 1 {
		host += ":21"
	}
	path = u.Path
	if u.User != nil {
		user = u.User.Username()
		pass, _ = u.User.Password()
	}
	return
}
