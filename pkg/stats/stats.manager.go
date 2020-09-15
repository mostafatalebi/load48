package stats

import (
	"fmt"
	dyanmic_params "github.com/mostafatalebi/dynamic-params"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	Total                = "total"
	CacheUsed            = "cache-used"
	Success              = "success"
	Timeout              = "timeout"
	FailedPrefix         = "failed::"
	Failed               = FailedPrefix + "%v"
	MainDuration         = "main-duration"
	ExecDuration         = "exec-duration"
	LongestDuration      = "longest-duration"
	AverageDuration      = "average-duration"
	ShortestDuration     = "shortest-duration"
	LongestExecDuration  = "longest-exec-duration"
	AverageExecDuration  = "average-exec-duration"
	ShortestExecDuration = "shortest-exec-duration"
)

var DefaultPreset = map[string]string{
	Total: "Total Number of Requests",
	Success: "Total Success",
	Timeout: "Total Timeouts",
	Failed+"::500": "Failed 500",
	Failed+"::501": "Failed 501",
	Failed+"::502": "Failed 502",
	Failed+"::404": "Failed 404",
	Failed+"::401": "Failed 401",
	Failed+"::403": "Failed 403",
	Failed+"::400": "Failed 400",
	AverageExecDuration: "Average App Execution",
	AverageDuration: "Average Duration",
	ShortestDuration: "Shortest Duration",
	ShortestExecDuration: "Shortest App Execution",
	LongestDuration: "Longest Duration",
	LongestExecDuration: "Longest App Execution",
}

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
		case Success:
			vv := value.(int64)
			s.IncrSuccess(vv)
		case CacheUsed:
			vv := value.(int64)
			s.IncrCacheUsed(vv)
		case Failed:
			vv := value.(int64)
			fc := strings.Replace(key, FailedPrefix, "", 1)
			fcode, _ := strconv.Atoi(fc)
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
}
