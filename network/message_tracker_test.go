package network_test

import (
	"fmt"
	"github.com/google/uuid"
	"testing"

	"github.com/ChainSafe/gossamer-go-interview/network"
	"github.com/stretchr/testify/assert"
)

func generateMessage(n int) *network.Message {
	return &network.Message{
		ID:     generateID(n),
		PeerID: fmt.Sprintf("somePeerID%d", n),
		Data:   []byte{0, 1, 1},
	}
}

func generateID(n int) string {
	return fmt.Sprintf("someID%d", n)
}

func generateMessageWithRandomUUID(n int) *network.Message {
	id := uuid.New()
	return &network.Message{
		ID:     id.String(),
		PeerID: fmt.Sprintf("somePeerID%d", n),
		Data:   []byte{0, 1, 1},
	}
}

func TestMessageTracker_Add(t *testing.T) {
	t.Run("add, get, then all messages", func(t *testing.T) {
		length := 5
		mt := network.NewMessageTracker(length)

		for i := 0; i < length; i++ {
			err := mt.Add(generateMessage(i))
			assert.NoError(t, err)

			msg, err := mt.Message(generateMessage(i).ID)
			assert.NoError(t, err)
			assert.NotNil(t, msg)
		}

		msgs := mt.Messages()
		assert.Equal(t, []*network.Message{
			generateMessage(0),
			generateMessage(1),
			generateMessage(2),
			generateMessage(3),
			generateMessage(4),
		}, msgs)
	})

	t.Run("add, get, then all messages, delete some", func(t *testing.T) {
		length := 5
		mt := network.NewMessageTracker(length)

		for i := 0; i < length; i++ {
			err := mt.Add(generateMessage(i))
			assert.NoError(t, err)

			msg, err := mt.Message(generateMessage(i).ID)
			assert.NoError(t, err)
			assert.NotNil(t, msg)
		}

		msgs := mt.Messages()
		assert.Equal(t, []*network.Message{
			generateMessage(0),
			generateMessage(1),
			generateMessage(2),
			generateMessage(3),
			generateMessage(4),
		}, msgs)

		for i := 0; i < length-2; i++ {
			err := mt.Delete(generateMessage(i).ID)
			assert.NoError(t, err)
		}

		msgs = mt.Messages()
		assert.Equal(t, []*network.Message{
			generateMessage(3),
			generateMessage(4),
		}, msgs)

	})

	t.Run("not full, with duplicates", func(t *testing.T) {
		length := 5
		mt := network.NewMessageTracker(length)

		for i := 0; i < length-1; i++ {
			err := mt.Add(generateMessage(i))
			assert.NoError(t, err)
		}
		for i := 0; i < length-1; i++ {
			err := mt.Add(generateMessage(length - 2))
			assert.NoError(t, err)
		}

		msgs := mt.Messages()
		assert.Equal(t, []*network.Message{
			generateMessage(0),
			generateMessage(1),
			generateMessage(2),
			generateMessage(3),
		}, msgs)
	})

	t.Run("not full, with duplicates from other peers", func(t *testing.T) {
		length := 5
		mt := network.NewMessageTracker(length)

		for i := 0; i < length-1; i++ {
			err := mt.Add(generateMessage(i))
			assert.NoError(t, err)
		}
		for i := 0; i < length-1; i++ {
			msg := generateMessage(length - 2)
			msg.PeerID = "somePeerID0"
			err := mt.Add(msg)
			assert.NoError(t, err)
		}

		msgs := mt.Messages()
		assert.Equal(t, []*network.Message{
			generateMessage(0),
			generateMessage(1),
			generateMessage(2),
			generateMessage(3),
		}, msgs)
	})
}

func TestMessageTracker_Cleanup(t *testing.T) {
	t.Run("overflow and cleanup", func(t *testing.T) {
		length := 5
		mt := network.NewMessageTracker(length)

		for i := 0; i < length*2; i++ {
			err := mt.Add(generateMessage(i))
			assert.NoError(t, err)
		}

		msgs := mt.Messages()
		assert.Equal(t, []*network.Message{
			generateMessage(5),
			generateMessage(6),
			generateMessage(7),
			generateMessage(8),
			generateMessage(9),
		}, msgs)
	})

	t.Run("overflow and cleanup with duplicate", func(t *testing.T) {
		length := 5
		mt := network.NewMessageTracker(length)

		for i := 0; i < length*2; i++ {
			err := mt.Add(generateMessage(i))
			assert.NoError(t, err)
		}

		for i := length; i < length*2; i++ {
			err := mt.Add(generateMessage(i))
			assert.NoError(t, err)
		}

		msgs := mt.Messages()
		assert.Equal(t, []*network.Message{
			generateMessage(5),
			generateMessage(6),
			generateMessage(7),
			generateMessage(8),
			generateMessage(9),
		}, msgs)
	})
}

func TestMessageTracker_Delete(t *testing.T) {
	t.Run("empty tracker", func(t *testing.T) {
		length := 5
		mt := network.NewMessageTracker(length)
		err := mt.Delete("bleh")
		assert.ErrorIs(t, err, network.ErrMessageNotFound)
	})
}

func TestMessageTracker_Message(t *testing.T) {
	t.Run("empty tracker", func(t *testing.T) {
		length := 5
		mt := network.NewMessageTracker(length)
		msg, err := mt.Message("bleh")
		assert.ErrorIs(t, err, network.ErrMessageNotFound)
		assert.Nil(t, msg)
	})
}

func TestMessageTracker_DeleteLastMessage(t *testing.T) {
	t.Run("delete last message", func(t *testing.T) {
		length := 5

		mt := network.NewMessageTracker(length)
		for i := 0; i < length; i++ {
			err := mt.Add(generateMessage(i))
			assert.NoError(t, err)
		}

		err := mt.Delete(generateID(4))
		assert.NoError(t, err)

		msgs := mt.Messages()
		assert.Equal(t, []*network.Message{
			generateMessage(0),
			generateMessage(1),
			generateMessage(2),
			generateMessage(3),
		}, msgs)
	})
}

func BenchmarkTestTrackerAddAndGetAllMessages_10000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		length := 10000
		mt := network.NewMessageTracker(length)

		for j := 0; j < length; j++ {
			_ = mt.Add(generateMessage(j))
		}

		_ = mt.Messages()
	}
}

func BenchmarkTestTrackerAddAndGetSpecificMessages_10000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		length := 10000
		mt := network.NewMessageTracker(length)

		for j := 0; j < length; j++ {
			_ = mt.Add(generateMessage(j))
			_, _ = mt.Message(generateID(j))
		}
	}
}

func BenchmarkTestTrackerOverflowGetAll_10000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		length := 10000
		mt := network.NewMessageTracker(length)

		for j := 0; j < length*2; j++ {
			_ = mt.Add(generateMessage(j))
		}

		_ = mt.Messages()
	}
}

func BenchmarkTestTrackerAddAndDeletingSome_10000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		length := 10000
		mt := network.NewMessageTracker(length)
		idsToDelete := make([]string, 0)

		for j := 0; j < length; j++ {
			msg := generateMessageWithRandomUUID(j)
			_ = mt.Add(msg)
			if j%20 == 0 {
				idsToDelete = append(idsToDelete, msg.ID)
			}
		}

		_ = mt.Messages()

		for _, id := range idsToDelete {
			err := mt.Delete(id)
			if err != nil {
				b.Fatal("wrong ID")
			}
		}

		_ = mt.Messages()
	}
}

func BenchmarkTestTrackerAddAndDeletingSome_100000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		length := 100000
		mt := network.NewMessageTracker(length)
		idsToDelete := make([]string, 0)

		for j := 0; j < length; j++ {
			msg := generateMessageWithRandomUUID(j)
			_ = mt.Add(msg)
			if j%20 == 0 {
				idsToDelete = append(idsToDelete, msg.ID)
			}
		}

		_ = mt.Messages()

		for _, id := range idsToDelete {
			err := mt.Delete(id)
			if err != nil {
				b.Fatal("wrong ID")
			}
		}

		_ = mt.Messages()
	}
}
