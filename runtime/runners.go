package runtime

import (
	"sync"
	"time"
)

type Runners struct {
	RWMutex            *sync.RWMutex
	QPSRWMutex         *sync.RWMutex
	StartingCountMutex *sync.RWMutex
	StartingCount      int

	QPS             map[int64]int
	CurrentPosition int       //轮询的指针
	Running         []*Runner `json:"instances"` //每个runner 可以有多个实例
}

func (r *Runners) RemoveRunner(uuid string) {
	r.RWMutex.Lock()
	defer r.RWMutex.Unlock()
	news := make([]*Runner, 0, len(r.Running))
	for _, runner := range r.Running {
		if runner.UUID != uuid {
			news = append(news, runner)
		}
	}
	r.Running = news
}

// GetOne  自动轮询取
func (r *Runners) GetOne() (rr *Runner, exit bool) {
	r.RWMutex.Lock()
	defer r.RWMutex.Unlock()
	if r.Running == nil || len(r.Running) == 0 {
		return nil, false
	}
	pos := r.CurrentPosition % len(r.Running)
	r.CurrentPosition += 1
	return r.Running[pos], true
}

func (r *Runners) GetRunningCount() int {
	return len(r.Running)
}

func (r *Runners) GetCurrentQPSCount() int {
	r.QPSRWMutex.Lock()
	defer r.QPSRWMutex.Unlock()
	return r.QPS[time.Now().Unix()]
}
func (r *Runners) AddCurrentQPS(count int) {
	r.QPSRWMutex.Lock()
	defer r.QPSRWMutex.Unlock()
	r.QPS[time.Now().Unix()] += count
}

func (r *Runners) AddStartingCount(count int) {
	r.StartingCountMutex.Lock()
	defer r.StartingCountMutex.Unlock()
	r.StartingCount += count
}
func (r *Runners) DelStartingCount(count int) {
	r.StartingCountMutex.Lock()
	defer r.StartingCountMutex.Unlock()
	r.StartingCount -= count
}
func (r *Runners) GetCurrentStartingCount() int {
	r.StartingCountMutex.Lock()
	defer r.StartingCountMutex.Unlock()
	return r.StartingCount
}
