package counters

import (
	"database/sql"
	"sync/atomic"

	"github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/query_gen"
)

var PostCounter *DefaultPostCounter

type DefaultPostCounter struct {
	buckets       [2]int64
	currentBucket int64

	insert *sql.Stmt
}

func NewPostCounter() (*DefaultPostCounter, error) {
	acc := qgen.NewAcc()
	counter := &DefaultPostCounter{
		currentBucket: 0,
		insert:        acc.Insert("postchunks").Columns("count, createdAt").Fields("?,UTC_TIMESTAMP()").Prepare(),
	}
	common.AddScheduledFifteenMinuteTask(counter.Tick)
	//common.AddScheduledSecondTask(counter.Tick)
	common.AddShutdownTask(counter.Tick)
	return counter, acc.FirstError()
}

func (counter *DefaultPostCounter) Tick() (err error) {
	var oldBucket = counter.currentBucket
	var nextBucket int64 // 0
	if counter.currentBucket == 0 {
		nextBucket = 1
	}
	atomic.AddInt64(&counter.buckets[oldBucket], counter.buckets[nextBucket])
	atomic.StoreInt64(&counter.buckets[nextBucket], 0)
	atomic.StoreInt64(&counter.currentBucket, nextBucket)

	var previousViewChunk = counter.buckets[oldBucket]
	atomic.AddInt64(&counter.buckets[oldBucket], -previousViewChunk)
	return counter.insertChunk(previousViewChunk)
}

func (counter *DefaultPostCounter) Bump() {
	atomic.AddInt64(&counter.buckets[counter.currentBucket], 1)
}

func (counter *DefaultPostCounter) insertChunk(count int64) error {
	if count == 0 {
		return nil
	}
	common.DebugLogf("Inserting a postchunk with a count of %d", count)
	_, err := counter.insert.Exec(count)
	return err
}
