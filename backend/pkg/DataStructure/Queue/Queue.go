package Queue

import (
	"sync"

	"github.com/faawibowo/Tubes2_Gopher/pkg/DataStructure/Graph"
	"github.com/faawibowo/Tubes2_Gopher/pkg/DataStructure/Tree"
)

type QueueItem struct {
	GraphNode *Graph.ElementGraph
	TreeNode  *Tree.TreeNodeElement
}

type Queue struct {
	items []*QueueItem
	mu    sync.Mutex
}

func NewQueue() *Queue {
	return &Queue{
		items: make([]*QueueItem, 0),
	}
}

func (q *Queue) Enqueue(item *QueueItem) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = append(q.items, item)
}

func (q *Queue) Dequeue() *QueueItem {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.items) == 0 {
		return nil
	}
	item := q.items[0]
	q.items = q.items[1:]
	return item
}

func (q *Queue) IsEmpty() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items) == 0
}

func (q *Queue) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}
