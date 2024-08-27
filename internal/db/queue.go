package db

import (
	"encoding/json"

	"github.com/Kasama/kasama-twitch-integrations/internal/logger"
	"github.com/beeker1121/goque"
)

type Queue[V any] struct {
	datadir string
	Backend *goque.PriorityQueue
}

func newQueue(datadir string) (*goque.PriorityQueue, error) {
	pq, err := goque.OpenPriorityQueue(datadir, goque.DESC)
	if err != nil {
		return nil, err
	}
	return pq, nil
}

func NewQueue[V any](datadir string) (*Queue[V], error) {
	pq, err := newQueue(datadir)
	if err != nil {
		return nil, err
	}
	q := &Queue[V]{
		datadir: datadir,
		Backend: pq,
	}

	return q, nil
}

func (q *Queue[V]) Clear() error {
	err := q.Backend.Drop()
	if err != nil {
		return err
	}
	backend, err := newQueue(q.datadir)
	if err != nil {
		return err
	}
	q.Backend = backend

	return nil
}

func (q *Queue[V]) Push(priority uint8, v V) {
	marshalled, err := json.Marshal(v)
	if err != nil {
		logger.Errorf("Failed to marshal type: %v", err)
	}
	_, err = q.Backend.Enqueue(priority, marshalled)
	if err != nil {
		logger.Errorf("Failed to enqueue: %v", err)
	}
}

func (q *Queue[V]) PushRaw(priority uint8, v []byte) {
	_, err := q.Backend.Enqueue(priority, v)
	if err != nil {
		logger.Errorf("Failed to enqueue: %v", err)
	}
}

func (q *Queue[V]) presentValue(value []byte) *V {
	var v V
	err := json.Unmarshal(value, &v)
	if err != nil {
		logger.Errorf("Failed to unmarshal type: %v", err)
		return nil
	}

	return &v
}

func (q *Queue[V]) Pop() *V {
	item, err := q.Backend.Dequeue()
	if err != nil {
		return nil
	}
	return q.presentValue(item.Value)
}

func (q *Queue[V]) Peek() *V {
	item, err := q.Backend.Peek()
	if err != nil {
		return nil
	}
	return q.presentValue(item.Value)
}

func (q *Queue[V]) Len() int {
	return int(q.Backend.Length())
}

func (q *Queue[V]) RawItems() []*goque.PriorityItem {
	items := make([]*goque.PriorityItem, 0, q.Len())

	var i uint64 = 0
	for {
		if i >= q.Backend.Length() {
			break
		}

		newItem, err := q.Backend.PeekByOffset(i)
		if err != nil {
			logger.Errorf("Failed to do something with queue: %v", err)
		}
		items = append(items, newItem)

		i++
	}

	return items
}

func (q *Queue[V]) Items() []V {
	items := make([]V, 0, q.Len())

	var i uint64 = 0
	for {
		if i >= q.Backend.Length() {
			break
		}

		newItem, err := q.Backend.PeekByOffset(i)
		if err != nil {
			logger.Errorf("Failed to do something with queue: %v", err)
		}
		value := q.presentValue(newItem.Value)
		items = append(items, *value)

		i++
	}

	return items
}
