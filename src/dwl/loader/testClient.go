package loader

import (
	"dwl/settings"
	"io"
	"math/rand"
	"time"
)

var allsize int64 = 10 * 1024 * 1024

type TestClient struct {
	pos, off int64
	wait     time.Duration
}

func NewTestClient(sets *settings.Settings, off int64) *TestClient {
	t := new(TestClient)
	t.off = off
	t.pos = off
	t.wait = time.Nanosecond * time.Duration(rand.Intn(5)+int(off/1000))
	return t
}

func (c *TestClient) Connect() error {
	//	time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000)))
	return nil
}
func (c *TestClient) Close() error {
	return nil
}
func (c *TestClient) Read(buf []byte) (n int, err error) {
	as := c.GetAllSize()

	if c.wait > 0 {
		time.Sleep(c.wait * time.Duration(len(buf)))
	}

	for i := 0; i < len(buf); i++ {
		buf[i] = byte(c.pos * 255 / as)
		c.pos++
		if c.pos >= as {
			return i, io.EOF
		}
	}
	return len(buf), nil
}
func (c *TestClient) IsConnected() bool {
	return true
}
func (c *TestClient) GetLastError() error {
	return nil
}
func (c *TestClient) GetAllSize() int64 {
	return allsize
}
func (c *TestClient) Pos() int64 {
	return c.pos
}
func (c *TestClient) Off() int64 {
	return c.off
}
