package progress

import (
	"fmt"
	"go.uber.org/atomic"
	"math"
	"sync"
)

type ProgressIndicator struct {
	PercentageCovered map[int8]int8
	Lock              *sync.Mutex
	Total             int64
	listenIncr        atomic.Int64
}

func NewProgressIndicator(total int64) *ProgressIndicator {
	return &ProgressIndicator{
		PercentageCovered: make(map[int8]int8, 10),
		Lock:              &sync.Mutex{},
		Total:             total,
	}
}

func (p *ProgressIndicator) ByPercent(total, current int64, fn func(percent int8)) {
	if total < 10 {
		return
	}
	p.Lock.Lock()
	defer p.Lock.Unlock()

	var each10Percent = float64(total) / float64(10)
	var cp int8
	if float64(current) < each10Percent {
		cp = int8(0)
	} else if current == total {
		cp = int8(100)
	} else {
		remainder := math.Mod(float64(current), each10Percent)
		if remainder == 0 {
			cp = int8(float64(current) / each10Percent)
			cp = cp * 10
		} else {
			var floorDivisible = float64(current) - remainder
			cp = int8(floorDivisible / each10Percent)
			cp = cp * 10
		}
	}
	if _, ok := p.PercentageCovered[cp]; !ok {
		p.PercentageCovered[cp] = cp
		fn(cp)
	}
}

func (p *ProgressIndicator) Print(current int64) {
	p.ByPercent(p.Total, current, func(percent int8) {
		if percent == 0 {
			return
		} else if percent == int8(100) {
			fmt.Printf("==%v%v [%v completed!]", "%", percent, p.Total)
		} else {
			fmt.Printf("==%v%v==", "%", percent)
		}
	})
}

func (p *ProgressIndicator) ListenToChannel(ch chan int8) {
	for _ = range ch {
		p.listenIncr.Add(1)
		p.Print(p.listenIncr.Load())
	}
}
