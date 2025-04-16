package runtime

import (
	"github.com/yunhanshu-net/runcher/runner"
	"sync"
	"time"
)

func NewRunners(running ...runner.Runner) *Runners {
	rs := &Runners{
		qpsLock:        &sync.Mutex{},
		getLock:        &sync.Mutex{},
		connectLock:    &sync.Mutex{},
		currentPos:     0,
		latestHandelTs: time.Now(),
		StartLock:      make(map[string]*sync.Mutex),
		QpsWindows:     make(map[int64]int64),
		Running:        make([]runner.Runner, 0, 2),
	}
	if len(running) > 0 {
		rs.Running = running
	}
	return rs
}

type Runners struct {
	qpsLock        *sync.Mutex
	getLock        *sync.Mutex
	connectLock    *sync.Mutex
	latestHandelTs time.Time
	currentPos     int

	StartLock map[string]*sync.Mutex

	QpsWindows map[int64]int64
	Running    []runner.Runner
}

func (r *Runners) AddQps(count int64) {
	r.qpsLock.Lock()
	r.latestHandelTs = time.Now()
	r.QpsWindows[time.Now().Unix()] += count
	r.qpsLock.Unlock()
}

func (r *Runners) GetCurrentQps() int64 {
	r.qpsLock.Lock()
	defer r.qpsLock.Unlock()
	return r.QpsWindows[time.Now().Unix()]
}

func (r *Runners) GetOne() runner.Runner {
	r.getLock.Lock()
	defer r.getLock.Unlock()
	if len(r.Running) == 0 {
		return nil
	}
	i := r.currentPos % len(r.Running)
	r.currentPos++
	return r.Running[i]
}
