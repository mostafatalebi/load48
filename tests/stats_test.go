package tests

import (
	"github.com/mostafatalebi/loadtest/pkg/stats"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestTimeOutCounter_mustBeTrue(t *testing.T){
	st := stats.NewStatsManager("test_1")
	st2 := stats.NewStatsManager("test_1")

	for i := 0; i < 100; i++ {
		st.IncrTimeout(1)
		st2.IncrTimeout(1)
	}

	v, err := st.Params.GetAsInt64("timeout")
	assert.NoError(t, err)
	assert.Equal(t, int64(100), v)
	v, err = st2.Params.GetAsInt64("timeout")
	assert.NoError(t, err)
	assert.Equal(t, int64(100), v)

	var stNew = st2.Merge(st)
	v, err = stNew.Params.GetAsInt64("timeout")
	assert.NoError(t, err)
	assert.Equal(t, int64(200), v)
}

func TestTimeOutCounterConcurrent_mustBeTrue(t *testing.T){
	st := stats.NewStatsManager("test_1")

	wg := &sync.WaitGroup{}

	for i := 0; i < 10000; i++ {
		wg.Add(3)
		go func() {
			st.IncrTimeout(1)
			wg.Done()
		}()
		go func() {
			st.IncrTimeout(1)
			wg.Done()
		}()
		go func() {
			st.IncrTimeout(1)
			wg.Done()
		}()
	}
	wg.Wait()
	v, err := st.Params.GetAsInt64("timeout")
	assert.NoError(t, err)
	assert.Equal(t, int64(10000*3), v)
}
