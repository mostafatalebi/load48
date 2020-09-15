package stats

import (
	"fmt"
	dyanmic_params "github.com/mostafatalebi/dynamic-params"
	"regexp"
	"strconv"
	"sync"
	"time"
)

const (
	Total                = "total"
	CacheUsed            = "cache-used"
	Success              = "success"
	Timeout              = "timeout"
	ConnRefused          = "connection-refused"
	OtherErrors          = "other-errors"
	Failed               = "%v"
	MainDuration         = "main-duration"
	ExecDuration         = "exec-duration"
	LongestDuration      = "longest-duration"
	AverageDuration      = "average-duration"
	ShortestDuration     = "shortest-duration"
	LongestExecDuration  = "longest-exec-duration"
	AverageExecDuration  = "average-exec-duration"
	ShortestExecDuration = "shortest-exec-duration"
)

type StatsCollector struct {
	lock   *sync.RWMutex
	Key    string
	Params *dyanmic_params.DynamicParams
}

//
func NewStatsManager(key string) *StatsCollector {
	params := dyanmic_params.NewDynamicParams(dyanmic_params.SrcNameInternal, &sync.RWMutex{})
	return &StatsCollector{
		lock:   &sync.RWMutex{},
		Key:    key,
		Params: params,
	}
}

func (s *StatsCollector) IncrSuccess(incr int64) {
	s.lock.Lock()
	defer s.lock.Unlock()
	v, err := s.Params.GetAsInt64(Success)
	if err != nil && err.Error() != dyanmic_params.ErrNotFound {
		return
	} else if err != nil && err.Error() == dyanmic_params.ErrNotFound {
		s.Params.Add(Success, incr)
		return
	}
	s.Params.Add(Success, v+incr)
}

func (s *StatsCollector) IncrCacheUsed(incr int64) {
	s.lock.Lock()
	defer s.lock.Unlock()
	v, err := s.Params.GetAsInt64(CacheUsed)
	if err != nil && err.Error() != dyanmic_params.ErrNotFound {
		return
	} else if err != nil && err.Error() == dyanmic_params.ErrNotFound {
		s.Params.Add(CacheUsed, incr)
		return
	}
	s.Params.Add(CacheUsed, v+incr)
}

func (s *StatsCollector) IncrTimeout(incr int64) {
	s.lock.Lock()
	defer s.lock.Unlock()
	v, err := s.Params.GetAsInt64(Timeout)
	if err != nil && err.Error() != dyanmic_params.ErrNotFound {
		return
	} else if err != nil && err.Error() == dyanmic_params.ErrNotFound {
		s.Params.Add(Timeout, incr)
		return
	}
	s.Params.Add(Timeout, v+incr)
}

func (s *StatsCollector) IncrConnRefused(incr int64) {
	s.lock.Lock()
	defer s.lock.Unlock()
	v, err := s.Params.GetAsInt64(ConnRefused)
	if err != nil && err.Error() != dyanmic_params.ErrNotFound {
		return
	} else if err != nil && err.Error() == dyanmic_params.ErrNotFound {
		s.Params.Add(ConnRefused, incr)
		return
	}
	s.Params.Add(ConnRefused, v+incr)
}

func (s *StatsCollector) IncrOtherErrors(incr int64) {
	s.lock.Lock()
	defer s.lock.Unlock()
	v, err := s.Params.GetAsInt64(OtherErrors)
	if err != nil && err.Error() != dyanmic_params.ErrNotFound {
		return
	} else if err != nil && err.Error() == dyanmic_params.ErrNotFound {
		s.Params.Add(OtherErrors, incr)
		return
	}
	s.Params.Add(OtherErrors, v+incr)
}

func (s *StatsCollector) IncrTotal(incr int64) {
	s.lock.Lock()
	defer s.lock.Unlock()
	v, err := s.Params.GetAsInt64(Total)
	if err != nil && err.Error() != dyanmic_params.ErrNotFound {
		return
	} else if err != nil && err.Error() == dyanmic_params.ErrNotFound {
		s.Params.Add(Total, incr)
		return
	}
	s.Params.Add(Total, v+incr)
}

func (s *StatsCollector) IncrFailed(failureCode int, incr int64) {
	s.lock.Lock()
	defer s.lock.Unlock()
	v, err := s.Params.GetAsInt64(fmt.Sprintf(Failed, failureCode))
	if err != nil && err.Error() != dyanmic_params.ErrNotFound {
		return
	} else if err != nil && err.Error() == dyanmic_params.ErrNotFound {
		s.Params.Add(fmt.Sprintf(Failed, failureCode), incr)
		return
	}
	s.Params.Add(fmt.Sprintf(Failed, failureCode), v+incr)
}

func (s *StatsCollector) AddMainDuration(duration time.Duration) {
	s.lock.Lock()
	defer s.lock.Unlock()
	r, err := s.Params.GetAsTimeDuration(MainDuration)
	if err != nil && err.Error() != dyanmic_params.ErrNotFound {
		return
	} else if err != nil && err.Error() == dyanmic_params.ErrNotFound {
		s.Params.Add(MainDuration, duration)
		return
	}
	newTimeDur := *r + duration
	s.Params.Add(MainDuration, newTimeDur)
}

func (s *StatsCollector) AddExecDuration(duration time.Duration) {
	s.lock.Lock()
	defer s.lock.Unlock()
	r, err := s.Params.GetAsTimeDuration(ExecDuration)
	if err != nil && err.Error() != dyanmic_params.ErrNotFound {
		return
	} else if err != nil && err.Error() == dyanmic_params.ErrNotFound {
		s.Params.Add(ExecDuration, duration)
		return
	}
	newTimeDur := *r + duration
	s.Params.Add(ExecDuration, newTimeDur)
}

func (s *StatsCollector) AddLongestDuration(duration time.Duration) {
	s.lock.Lock()
	defer s.lock.Unlock()
	r, err := s.Params.GetAsTimeDuration(LongestDuration)
	if err != nil && err.Error() != dyanmic_params.ErrNotFound {
		return
	} else if err != nil && err.Error() == dyanmic_params.ErrNotFound {
		s.Params.Add(LongestDuration, duration)
		return
	}

	if *r < duration {
		s.Params.Add(LongestDuration, duration)
	}
}

func (s *StatsCollector) AddShortestDuration(duration time.Duration) {
	s.lock.Lock()
	defer s.lock.Unlock()
	r, err := s.Params.GetAsTimeDuration(ShortestDuration)
	if err != nil && err.Error() != dyanmic_params.ErrNotFound {
		return
	} else if err != nil && err.Error() == dyanmic_params.ErrNotFound {
		s.Params.Add(ShortestDuration, duration)
		return
	}

	if *r > duration {
		s.Params.Add(ShortestDuration, duration)
	}
}

func (s *StatsCollector) AddAverageDuration() {
	s.lock.Lock()
	defer s.lock.Unlock()
	rTotal, err := s.Params.GetAsInt64(Total)
	if err != nil && err.Error() != dyanmic_params.ErrNotFound {
		return
	} else if rTotal == 0 {
		return
	}
	rDur, err := s.Params.GetAsTimeDuration(MainDuration)
	if err != nil && err.Error() != dyanmic_params.ErrNotFound {
		return
	}
	duration := time.Duration(rDur.Nanoseconds() / rTotal)
	s.Params.Add(AverageDuration, duration)
}

func (s *StatsCollector) AddExecLongestDuration(duration time.Duration) {
	s.lock.Lock()
	defer s.lock.Unlock()
	r, err := s.Params.GetAsTimeDuration(LongestExecDuration)
	if err != nil && err.Error() != dyanmic_params.ErrNotFound {
		return
	} else if err != nil && err.Error() == dyanmic_params.ErrNotFound {
		s.Params.Add(LongestExecDuration, duration)
		return
	}

	if *r == 0 || *r < duration {
		s.Params.Add(LongestExecDuration, duration)
	}
}

func (s *StatsCollector) AddExecShortestDuration(duration time.Duration) {
	s.lock.Lock()
	defer s.lock.Unlock()
	r, err := s.Params.GetAsTimeDuration(ShortestExecDuration)
	if err != nil && err.Error() != dyanmic_params.ErrNotFound {
		return
	} else if err != nil && err.Error() == dyanmic_params.ErrNotFound {
		s.Params.Add(ShortestExecDuration, duration)
		return
	}

	if *r == 0 || *r > duration {
		s.Params.Add(ShortestExecDuration, duration)
	}
}

func (s *StatsCollector) AddExecAverageDuration() {
	s.lock.Lock()
	defer s.lock.Unlock()
	rTotal, err := s.Params.GetAsInt64(Total)
	if err != nil && err.Error() != dyanmic_params.ErrNotFound {
		return
	} else if rTotal == 0 {
		return
	}
	rDur, err := s.Params.GetAsTimeDuration(ExecDuration)
	if err != nil && err.Error() != dyanmic_params.ErrNotFound {
		return
	}
	duration := time.Duration(rDur.Nanoseconds() / rTotal)
	s.Params.Add(AverageExecDuration, duration)
}

func (s *StatsCollector) Merge(scp *StatsCollector) StatsCollector {
	s.Params.Iterate(func(key string, origValue interface{}) {
		if !scp.Params.Has(key) {
			return
		}
		value := scp.Params.Get(key)
		switch key {
		case Total:
			vv := value.(int64)
			s.IncrTotal(vv)
		case Timeout:
			vv := value.(int64)
			s.IncrTimeout(vv)
		case ConnRefused:
			vv := value.(int64)
			s.IncrConnRefused(vv)
		case OtherErrors:
			vv := value.(int64)
			s.IncrOtherErrors(vv)
		case Success:
			vv := value.(int64)
			s.IncrSuccess(vv)
		case CacheUsed:
			vv := value.(int64)
			s.IncrCacheUsed(vv)
		case Failed:
			vv := value.(int64)
			fcode, _ := strconv.Atoi(key)
			s.IncrFailed(fcode, vv)
		case ExecDuration:
			vv := value.(time.Duration)
			s.AddExecDuration(vv)
		case MainDuration:
			vv := value.(time.Duration)
			s.AddMainDuration(vv)
		case ShortestDuration:
			vv := value.(time.Duration)
			s.AddShortestDuration(vv)
		case AverageDuration:
			s.AddAverageDuration()
		case AverageExecDuration:
			s.AddExecAverageDuration()
		}
	})

	return *s
}

func (s *StatsCollector) PrintPretty(preset map[string]string) {
	fmt.Println("\n======== " + s.Key + " ========")
	for k, v := range preset {
		if s.Params.Has(k) {
			fmt.Printf("--- %v => %v \n", v, s.Params.Get(k))
		}
	}
	s.Params.Iterate(func(key string, value interface{}) {
		val, ok := value.(int64)
		if !ok {
			return
		}
		if m, err := regexp.Match(`^[0-9]+$`, []byte(key)); err == nil && m {
			fmt.Printf("--- Total Failed(%v) => %v \n", key, val)
		}
	})
}
