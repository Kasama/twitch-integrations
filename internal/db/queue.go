package db

import (
	"encoding/json"

	"github.com/Kasama/kasama-twitch-integrations/internal/logger"
	"github.com/beeker1121/goque"
)

type Queue[V any] struct {
	Backend *goque.PriorityQueue
}

func NewQueue[V any](datadir string) (*Queue[V], error) {
	pq, err := goque.OpenPriorityQueue(datadir, goque.DESC)
	if err != nil {
		return nil, err
	}
	q := &Queue[V]{
		Backend: pq,
	}

	return q, nil
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

func (q *Queue[V]) presentValue(value []byte) *V{
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
