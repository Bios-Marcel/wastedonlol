package main

import (
	"sync"
	"time"
)

type Limiter struct {
	triesLeft int
	locked    bool
	lock      *sync.Mutex
}

func NewLimiter(tries int, per time.Duration) *Limiter {
	limiter := &Limiter{
		triesLeft: tries,
		lock:      &sync.Mutex{},
	}

	resetTicker := time.NewTicker(per)
	go func() {
		for {
			<-resetTicker.C
			limiter.triesLeft = tries
			if limiter.locked {
				limiter.locked = false
				limiter.lock.Unlock()
			}
		}
	}()

	return limiter
}

func (limiter *Limiter) Wait() {
	limiter.lock.Lock()
	limiter.locked = true
	defer func() {
		limiter.locked = false
		limiter.lock.Unlock()
	}()
	if limiter.triesLeft <= 0 {
		limiter.lock.Lock()
	}

	limiter.triesLeft--
}
