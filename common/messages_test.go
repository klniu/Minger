package common

import (
	"testing"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type CommonSuite struct {}

var _ = Suite(&CommonSuite{})

func (s *CommonSuite) TestMessageQueue(c *C) {
	queue := NewMessageQueue(2)
	c.Assert(queue.IsEmpty(), Equals, true)
	mes := &Message{Format: "str", Content: []byte("test1")}
	queue.Push(mes)
	c.Assert(queue.IsEmpty(), Equals, false)
	c.Assert(queue.Pop(), Equals, mes)
	c.Assert(queue.IsEmpty(), Equals, true)

	queue.Push(mes)
	queue.Push(&Message{Format: "str", Content: []byte("test2")})
	queue.Push(&Message{Format: "str", Content: []byte("test3")})

	c.Assert(queue.Size(), Equals, 3)
	queue.Pop()
	queue.Pop()
	queue.Pop()
	c.Assert(queue.Size(), Equals, 0)
}
