package network

import (
	"errors"
	"runtime"
)

// MessageTracker tracks a configurable fixed amount of messages.
// Messages are stored first-in-first-out. Duplicate messages should not be stored in the queue.
type MessageTracker interface {
	// Add will add a message to the tracker, deleting the oldest message if necessary
	Add(message *Message) (err error)
	// Delete will delete message from tracker
	Delete(id string) (err error)
	// Message returns a message for a given ID. Message is retained in tracker
	Message(id string) (message *Message, err error)
	// Messages returns messages in FIFO order
	Messages() (messages []*Message)
}

// Tracker represents all the fields needed for the implementation.
type Tracker struct {
	msgMap  map[string]*Message
	msgList []*Message
	length  int
}

// asserting that Tracker struct implements interface MessageTracker.
var _ MessageTracker = &Tracker{}

// ErrMessageNotFound is an error returned by MessageTracker when a message with specified id is not found
var ErrMessageNotFound = errors.New("message not found")

// NewMessageTracker creates Tracker
func NewMessageTracker(length int) MessageTracker {
	return &Tracker{
		msgMap:  make(map[string]*Message),
		msgList: make([]*Message, 0),
		length:  length,
	}
}

// Add MessageTracker implementation: handling duplicates, full Tracker and normal addition.
func (t *Tracker) Add(message *Message) error {
	if t.messageExists(message.ID) {
		return nil
	}

	if t.isTrackerFull() {
		delete(t.msgMap, t.msgList[0].ID)
		t.msgList = t.msgList[1:]
	}

	t.msgMap[message.ID] = message
	t.msgList = append(t.msgList, message)

	return nil
}

// Delete MessageTracker implementation: deleting the Message by ID from map and the list.
func (t *Tracker) Delete(id string) error {
	if !t.messageExists(id) {
		return ErrMessageNotFound
	}

	msgIndex := t.getMessageIndex(id)
	t.msgList = append(t.msgList[:msgIndex], t.msgList[msgIndex+1:]...)
	delete(t.msgMap, id)

	return nil
}

// Message MessageTracker implementation: getting the Message by ID from the map.
func (t *Tracker) Message(id string) (*Message, error) {
	msg, exists := t.msgMap[id]
	if !exists {
		return nil, ErrMessageNotFound
	}

	return msg, nil
}

// Messages MessageTracker implementation: returning all the Message FIFO ordered
func (t *Tracker) Messages() []*Message {
	return t.msgList
}

func (t *Tracker) isTrackerFull() bool {
	return len(t.msgMap) >= t.length
}

func (t *Tracker) messageExists(id string) bool {
	_, exists := t.msgMap[id]
	return exists
}

func getMessageIndexByBatch(id string, list []*Message, from int, indexChan chan<- int) {
	for i := 0; i < len(list); i++ {
		if list[i].ID == id {
			indexChan <- from + i
			return
		}
	}
}

func (t *Tracker) getMessageIndex(id string) int {
	// uses NumCPU as amount of threads
	threads := runtime.NumCPU()
	msgLength := len(t.msgList)
	indexChan := make(chan int)
	defer close(indexChan)

	// if there is fewer messages that goroutines, then only spawn 1 threads
	batch := msgLength / threads
	if batch == 0 {
		threads = 1
		batch = msgLength
	}

	// splits in batches for the linear search
	for i := 0; i < threads; i++ {
		start := i * batch
		end := start + batch
		if i == threads-1 {
			end = msgLength
		}

		go getMessageIndexByBatch(id, t.msgList[start:end], start, indexChan)
	}

	msgIndex := <-indexChan
	return msgIndex
}
