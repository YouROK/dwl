package ftpClient

import (
	"dwl/settings"
	"errors"
	"net"
	"sync"
	"time"

	ftp "github.com/jum/tinyftp"
)

type FTP struct {
	conn  *ftp.Conn
	dconn net.Conn

	opt  *settings.Settings
	path string
	off  int64
	pos  int64
	size int64

	mutex sync.Mutex
	err   error
}

func NewFTP(opt *settings.Settings, off int64) *FTP {
	f := new(FTP)
	f.off = off
	f.pos = off
	f.opt = opt
	return f
}

func (f *FTP) Connect() error {

	if f.conn != nil {
		return nil
	}

	f.mutex.Lock()
	defer f.mutex.Unlock()

	host, path, user, pass, err := parseUrl(f.opt.Url)

	if err != nil {
		f.err = err
		return err
	}

	if path == "" || path == "/" {
		f.err = errors.New("wrong file path")
		return f.err
	}

	f.path = path
	timeout := f.opt.Config.GetInt("Timeout")

	if timeout == 0 {
		f.conn, _, _, err = ftp.Dial("tcp", host)
	} else {
		f.conn, _, _, err = ftp.DialTimeout("tcp", host, time.Millisecond*time.Duration(timeout))
	}

	if err != nil {
		return err
	}

	if user == "" {
		user = f.opt.Config.GetStr("user")
	}

	if pass == "" {
		pass = f.opt.Config.GetStr("pass")
	}

	_, _, err = f.conn.Login(user, pass)
	if err != nil {
		f.conn.Close()
		f.conn = nil
		f.err = err
		return err
	}

	_, _, err = f.conn.Type("I")
	if err != nil {
		f.conn.Close()
		f.conn = nil
		f.err = err
		return err
	}

	s, _, _, err := f.conn.Size(f.path)
	if err != nil {
		f.conn.Close()
		f.conn = nil
		f.err = err
		return err
	}
	f.size = s

	_, _, err = f.conn.Rest(f.pos)
	if err != nil {
		f.conn.Close()
		f.conn = nil
		f.err = err
		return err
	}

	addr, _, _, err := f.conn.Passive()
	if err != nil {
		f.conn.Close()
		f.conn = nil
		f.err = err
		return err
	}

	f.dconn, err = net.Dial("tcp", addr)
	if err != nil {
		f.conn.Close()
		f.conn = nil
		f.err = err
		return err
	}

	_, _, err = f.conn.Cmd(1, "RETR %s", path)
	if err != nil {
		f.conn.Close()
		f.conn = nil
		f.err = err
		return err
	}

	return nil
}

func (f *FTP) IsConnected() bool {
	return f.conn != nil
}

func (f *FTP) GetLastError() error {
	return f.err
}

func (f *FTP) GetAllSize() int64 {
	return f.size
}

func (f *FTP) Off() int64 {
	return f.off
}

func (f *FTP) Pos() int64 {
	return f.pos
}

func (f *FTP) Read(buf []byte) (n int, err error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.conn != nil && f.dconn != nil {
		n, err = f.dconn.Read(buf)
		if n > 0 {
			f.pos += int64(n)
		}
		if err != nil {
			f.err = err
		}
		return
	}
	f.err = errors.New("read after close")
	return 0, f.err
}

func (f *FTP) Close() error {
	if f == nil {
		return nil
	}

	f.mutex.Lock()
	defer f.mutex.Unlock()
	var err error
	if f.dconn != nil {
		err = f.dconn.Close()
		f.dconn = nil
	}

	if f.conn != nil {
		err = f.conn.Close()
		f.conn = nil
	}
	return err
}
