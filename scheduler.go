package main

import (
	"sync"
	"time"
)

type Scheduler struct {
	mu    sync.Mutex
	batch map[string]time.Time
	done  chan bool
}

func (s *Scheduler) ProcessBatch(db *Database) {
	s.mu.Lock()
	defer s.mu.Unlock()
	db.Update(s.batch)
	s.batch = make(map[string]time.Time)
}

func (s *Scheduler) Stop() {
	s.done <- true
}

func (s *Scheduler) Run(db *Database) {
	for routine_pool := 0; routine_pool < ROUTINE_POOL_MAX; routine_pool++ {
		go func() {
			limiter := time.Tick(500 * time.Millisecond)
			for {
				select {
				case <-limiter:
					s.ProcessBatch(db)
				case <-s.done:
					break
				}
			}
		}()
	}
}

func (s *Scheduler) Add(resource_id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.batch[resource_id] = time.Time{} //zero time to differentiate between processed and not
}
