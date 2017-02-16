package httpClient

import (
	"dwl/settings"
	"errors"
	"net"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Http struct {
	client *http.Client
	req    *http.Request
	resp   *http.Response
	opt    *settings.Settings

	off int64
	pos int64

	mutex sync.Mutex
	err   error
}

func NewHttp(opt *settings.Settings, off int64) *Http {
	con := new(Http)
	con.off = off
	con.pos = off
	con.opt = opt
	return con
}

func (h *Http) Connect() error {
	h.mutex.Lock()
	h.client = &http.Client{}
	h.req, h.err = http.NewRequest("GET", h.opt.Url, nil)
	h.mutex.Unlock()
	if h.err != nil {
		defer h.Close()
		return h.err
	}

	h.req.Header.Set("Accept", "*/*")
	h.req.Header.Set("UserAgent", "DWL/1.0.0 ("+runtime.GOOS+")")

	if header := h.opt.Config.GetCfg("Header"); header != nil {
		for k, _ := range header {
			v := header.GetStr(k)
			h.req.Header.Add(k, v)
		}
	}
	h.req.Header.Add("Range", "bytes="+strconv.FormatInt(h.off, 10)+"-")

	h.mutex.Lock()
	timeout := h.opt.Config.GetInt("Timeout")
	if timeout > 0 {
		h.client.Timeout = time.Millisecond * time.Duration(timeout)
		var netTransport = &http.Transport{
			Dial: (&net.Dialer{
				Timeout: h.client.Timeout,
			}).Dial,
			TLSHandshakeTimeout:   h.client.Timeout,
			ResponseHeaderTimeout: h.client.Timeout,
			ExpectContinueTimeout: h.client.Timeout,
			IdleConnTimeout:       h.client.Timeout,
		}
		h.client.Transport = netTransport
	}
	h.resp, h.err = h.client.Do(h.req)
	if h.resp != nil && (h.resp.StatusCode != http.StatusOK && h.resp.StatusCode != http.StatusPartialContent) {
		h.err = errors.New(h.resp.Status)
	}
	h.mutex.Unlock()

	if h.err != nil {
		h.Close()
		return h.err
	}
	if h.resp != nil && h.resp.Request != nil && h.resp.Request.URL != nil {
		h.opt.Url = h.resp.Request.URL.String()
	}
	return nil
}

func (h *Http) IsConnected() bool {
	return h.resp != nil && h.resp.Body != nil && !h.resp.Close
}

func (h *Http) GetLastError() error {
	return h.err
}

func (h *Http) Read(buf []byte) (n int, err error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.resp != nil && h.resp.Body != nil {
		n, err = h.resp.Body.Read(buf)
		if n > 0 {
			h.pos += int64(n)
		}
		return
	}
	h.err = http.ErrBodyReadAfterClose
	return 0, http.ErrBodyReadAfterClose
}

func (h *Http) Close() error {
	if h == nil {
		return nil
	}

	h.mutex.Lock()
	defer h.mutex.Unlock()
	var err error
	if h.req != nil && h.req.Body != nil && !h.req.Close {
		err = h.req.Body.Close()
	}
	if h.resp != nil && h.resp.Body != nil && !h.resp.Close {
		err = h.resp.Body.Close()
	}
	h.req = nil
	h.resp = nil
	h.client = nil
	return err
}

func (h *Http) GetHeader(key string) string {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.resp != nil && h.resp.Header != nil {
		return h.resp.Header.Get(key)
	}
	return ""
}

func (h *Http) GetHeaders() http.Header {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.resp != nil && h.resp.Header != nil {
		return h.resp.Header
	}
	return nil
}

func (h *Http) GetAllSize() int64 {
	sizeStr := h.GetHeader("Content-Length")
	if sizeStr != "" {
		size, err := strconv.ParseInt(h.resp.Header["Content-Length"][0], 10, 0)
		if err == nil {
			return size
		}
		h.err = err
	} else if sizeStr = h.GetHeader("Content-Range"); sizeStr != "" {
		if cr := strings.Split(sizeStr, "/"); len(cr) > 0 {
			size, err := strconv.ParseInt(cr[len(cr)-1], 10, 0)
			if err == nil {
				return size
			}
			h.err = err
		}
	}
	return -1
}

func (h *Http) Off() int64 {
	return h.off
}

func (h *Http) Pos() int64 {
	return h.pos
}
