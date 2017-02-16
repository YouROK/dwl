package loader

import (
	"dwl/settings"
	"io"
	"math/rand"
	"time"
)

type TestClient struct {
	pos, off int64
	wait     time.Duration
}

func NewTestClient(sets *settings.Settings, off int64) *TestClient {
	t := new(TestClient)
	t.off = off
	t.pos = off
	t.wait = time.Nanosecond * time.Duration(rand.Intn(2000))
	//	fmt.Println("*** New client:", sets.Url, off)
	return t
}

func (c *TestClient) Connect() error {
	//	fmt.Println("Connect")
	time.Sleep(time.Millisecond * time.Duration(rand.Intn(2000)))
	return nil
}
func (c *TestClient) Close() error {
	//	fmt.Println("Close")
	return nil
}
func (c *TestClient) Read(buf []byte) (n int, err error) {
	//	fmt.Println("Read:", len(buf))
	//	defer fmt.Println("Readed:", n)
	allsize := c.GetAllSize()

	for i := 0; i < len(buf); i++ {
		time.Sleep(c.wait)
		buf[i] = byte(c.pos * 255 / allsize)
		c.pos++
		if c.pos >= allsize {
			return i, io.EOF
		}
	}
	return len(buf), nil
}
func (c *TestClient) IsConnected() bool {
	//	fmt.Println("Is connected")
	return true
}
func (c *TestClient) GetLastError() error {
	//	fmt.Println("Get last err")
	return nil
}
func (c *TestClient) GetAllSize() int64 {
	var allsize int64 = 5 * 1024 * 1024
	//	fmt.Println("Get all size:", allsize)
	return allsize
}
func (c *TestClient) Pos() int64 {
	//	fmt.Println("Pos:", c.pos)
	return c.pos
}
func (c *TestClient) Off() int64 {
	//	fmt.Println("Off:", c.off)
	return c.off
}
