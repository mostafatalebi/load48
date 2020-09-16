package tests

import (
	"github.com/gojektech/valkyrie"
	"github.com/mostafatalebi/loadtest/pkg/core"
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
			st.IncrTotal(1)
			wg.Done()
		}()
		go func() {
			st.IncrTimeout(1)
			st.IncrTotal(1)
			wg.Done()
		}()
		go func() {
			st.IncrTotal(1)
			wg.Done()
		}()
	}
	wg.Wait()
	v, err := st.Params.GetAsInt64(stats.Timeout)
	assert.NoError(t, err)
	assert.Equal(t, int64(10000*2), v)
	v, err = st.Params.GetAsInt64(stats.Total)
	assert.NoError(t, err)
	assert.Equal(t, int64(10000*3), v)
}

func TestMergingStats(t *testing.T){
	lt := core.NewAdGetLoadTest()
	st := stats.NewStatsManager("test_1")
	st2 := stats.NewStatsManager("test_2")
	st3 := stats.NewStatsManager("test_3")

	lt.AddStat("test_1", st)
	lt.AddStat("test_2", st2)
	lt.AddStat("test_3", st3)

	wg := &sync.WaitGroup{}

	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go func() {
			lt.GetStat("test_1").IncrTimeout(1)
			wg.Done()

		}()
		wg.Add(1)
		go func() {
			lt.GetStat("test_2").IncrTimeout(1)
			wg.Done()
		}()
		wg.Add(1)
		go func() {
			lt.GetStat("test_3").IncrTimeout(1)
			wg.Done()
		}()
	}

	wg.Wait()



	var stMerged = lt.MergeAll()

	v, err := stMerged.Params.GetAsInt64(stats.Timeout)
	assert.NoError(t, err)
	assert.Equal(t, int64(10000*3), v)
}

func TestErrorStrForFailedRequests(t *testing.T){
	lt := core.NewAdGetLoadTest()
	st := stats.NewStatsManager("test_1")

	lt.AddStat("test_1", st)

	wg := &sync.WaitGroup{}

	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go func() {
			err := &valkyrie.MultiError{}
			err.Push("http://example.com: context deadline exceeded (Client.Timeout exceeded while awaiting headers)")
			lt.UnderstandResponse("test_1", nil, err)
			wg.Done()
		}()
		wg.Add(1)
		go func() {
			err := &valkyrie.MultiError{}
			err.Push("Some unknown errors")
			lt.UnderstandResponse("test_1", nil, err)
			wg.Done()
		}()
	}

	wg.Wait()

	var stMerged = lt.MergeAll()

	v, err := stMerged.Params.GetAsInt64(stats.Timeout)
	assert.NoError(t, err)
	assert.Equal(t, int64(10000), v)
	v, err = stMerged.Params.GetAsInt64(stats.OtherErrors)
	assert.NoError(t, err)
	assert.Equal(t, int64(10000), v)
}
