package loader

import (
	"dwl/ftpClient"
	"dwl/httpClient"
	"dwl/settings"
	"errors"
	"net/url"
)

type Client interface {
	Connect() error
	Close() error
	Read(buf []byte) (n int, err error)
	IsConnected() bool
	GetLastError() error
	GetAllSize() int64
	Pos() int64
	Off() int64
}

func GetClient(sets *settings.Settings, pos int64) (Client, error) {
	url, err := url.Parse(sets.Url)
	if err != nil {
		return nil, err
	}
	switch url.Scheme {
	case "http", "https":
		return httpClient.NewHttp(sets, pos), nil
	case "ftp":
		return ftpClient.NewFTP(sets, pos), nil
	case "test":
		return NewTestClient(sets, pos), nil

	default:
		return nil, errors.New("not support protocol: " + url.Scheme)
	}
}
