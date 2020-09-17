package curr

import (
	"errors"
	"time"
)

const (
	MaxWaitError = "maximum wait limit passed"
	BadArgs = "backoff and interval values cannot be greater than maxWait value"
)

type Wait struct {
	interval time.Duration
	backoff time.Duration
	maxWait time.Duration
	itr int64
	err error
	stopChan *chan bool
}


func NewWait(interval, backoff, maxWait time.Duration) *Wait {
	if interval > maxWait || backoff > maxWait {
		return nil
	}
	return &Wait{
		interval: interval,
		maxWait: maxWait,
		backoff: backoff,
	}
}

func (w *Wait) SetChan(ch *chan bool){
	w.stopChan = ch
}

func (w *Wait) Waiting() bool {
	if w.itr > 0 {
		w.interval += w.backoff
	}

	if w.maxWait < w.interval {
		 w.err = errors.New(MaxWaitError)
		 return false
	}

	if w.stopChan != nil {
		go func() {

		}()
	}

	time.Sleep(w.interval)
	w.err = nil
	w.itr++
	return true
}

func (w *Wait) Error() error {
	return w.err
}
