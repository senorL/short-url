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
}

type LeafNode struct {
	current *Segment
	db      *gorm.DB
	mu      sync.Mutex
}

func fetchNextSegment(db *gorm.DB) (*Segment, error) {
	var gen model.IDGenerator
	if err := db.First(&gen, 1).Error; err != nil {
		return nil, err
	}

	newSegment := Segment{
		MaxID: gen.MaxID + uint64(gen.Step),
		NowID: gen.MaxID,
	}

	if err := db.Model(&gen).Update("max_id", newSegment.MaxID).Error; err != nil {
		return nil, err
	}
	return &newSegment, nil
}

func (l *LeafNode) GetID() (uint64, error) {
	id := atomic.AddUint64(&l.current.NowID, 1)

	if id <= l.current.MaxID {
		return id, nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if atomic.LoadUint64(&l.current.NowID) <= l.current.MaxID {
		return l.GetID()
	}
	newSegment, err := fetchNextSegment(l.db)
	if err != nil {
		return 0, err
	}
	l.current = newSegment

	return l.GetID()
}

func NewLeafNode(db *gorm.DB) (*LeafNode, error) {
	node := &LeafNode{
		db: db,
	}
	current, err := fetchNextSegment(db)
	if err != nil {
		return nil, err
	}
	node.current = current
	return node, nil
}
