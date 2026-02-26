package service

import (
	"short-url/internal/model"
	"sync"
	"sync/atomic"

	"gorm.io/gorm"
)

type Segment struct {
	MaxID uint64
	NowID uint64
	Step  int
}

type LeafNode struct {
	current      *Segment
	db           *gorm.DB
	mu           sync.Mutex
	isLoading    bool
	prefetchChan chan *Segment
}

func fetchNextSegment(db *gorm.DB) (*Segment, error) {
	var gen model.IDGenerator
	if err := db.First(&gen, 1).Error; err != nil {
		return nil, err
	}

	newSegment := Segment{
		MaxID: gen.MaxID + uint64(gen.Step),
		NowID: gen.MaxID,
		Step:  gen.Step,
	}

	if err := db.Model(&gen).Update("max_id", newSegment.MaxID).Error; err != nil {
		return nil, err
	}
	return &newSegment, nil
}

func (l *LeafNode) GetID() (uint64, error) {
	id := atomic.AddUint64(&l.current.NowID, 1)
	idMaxLimit := l.current.MaxID - uint64(float64(l.current.Step)*0.2)

	switch {
	case id <= idMaxLimit:
		return id, nil
	case idMaxLimit < id && id <= l.current.MaxID:
		l.asyncFetchNext()
		return id, nil
	default: // id > MaxID
		l.mu.Lock()
		defer l.mu.Unlock()

		if atomic.LoadUint64(&l.current.NowID) < l.current.MaxID {
			return l.GetID()
		}

		newSegment := <-l.prefetchChan
		l.current = newSegment
		l.isLoading = false
		return l.GetID()
	}
}

func (l *LeafNode) asyncFetchNext() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.isLoading {
		return
	}
	l.isLoading = true
	go func() {
		newSegment, err := fetchNextSegment(l.db)
		if err != nil {
			l.mu.Lock()
			l.isLoading = false
			l.mu.Unlock()
			return
		}
		l.prefetchChan <- newSegment
		return
	}()
}

func NewLeafNode(db *gorm.DB) (*LeafNode, error) {
	node := &LeafNode{
		db:           db,
		prefetchChan: make(chan *Segment, 1),
	}
	current, err := fetchNextSegment(db)
	if err != nil {
		return nil, err
	}
	node.current = current
	return node, nil
}
