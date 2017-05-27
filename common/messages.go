package common

type Message struct {
	User       string
	CreateTime int64
	Format     string
	Content    []byte
}

// NewQueue returns a new queue with the given initial size.
func NewMessageQueue(size int) *MessageQueue {
	return &MessageQueue{
		messages: make([]*Message, size),
		size:     size,
	}
}

// MessageQueue is a basic FIFO queue based on a circular list that resizes as needed.
type MessageQueue struct {
	messages []*Message
	size     int
	head     int
	tail     int
	count    int
	//handler  func()
}

// Push adds a node to the queue.
func (q *MessageQueue) Push(n *Message) {
	if q.head == q.tail && q.count > 0 {
		nodes := make([]*Message, len(q.messages) + q.size)
		copy(nodes, q.messages[q.head:])
		copy(nodes[len(q.messages) - q.head:], q.messages[:q.head])
		q.head = 0
		q.tail = len(q.messages)
		q.messages = nodes
	}
	q.messages[q.tail] = n
	q.tail = (q.tail + 1) % len(q.messages)
	q.count++
	//q.handler()
}

// IsEmpty returns whether the queue is empty.
func (q *MessageQueue) IsEmpty() bool {
	if q.count == 0 {
		return true
	}
	return false
}

// Size returns the size of the queue.
func (q *MessageQueue) Size() int {
	return q.count
}

// Pop removes and returns a node from the queue in first to last order.
func (q *MessageQueue) Pop() *Message {
	if q.count == 0 {
		return nil
	}
	node := q.messages[q.head]
	q.head = (q.head + 1) % len(q.messages)
	q.count--
	return node
}

// SetHandler set a function handling for handle the queue.
//func (q *MessageQueue) SetHandler(f func()) {
//	q.handler = f
//}
