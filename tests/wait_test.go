package tests

import (
	"github.com/mostafatalebi/loadtest/pkg/curr"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestWaitWithWrongParams(t *testing.T){
	w := curr.NewWait(time.Millisecond*10, time.Millisecond*5, time.Millisecond*1)
	assert.Nil(t, w)
}

func TestWaitWithBackoff(t *testing.T){
	w := curr.NewWait(time.Millisecond*10, time.Millisecond*5, time.Millisecond*21)
	itr := 0
	for w.Waiting() {
		itr++
	}
	assert.Equal(t, 3, itr)
}

func TestWaitWithBackoff_WithStopChannel(t *testing.T){
	w := curr.NewWait(time.Nanosecond*10, time.Nanosecond*5, time.Hour*1)
	var stopChan = make(chan bool)
	w.SetChan(&stopChan)
	itr := 0
	for w.Waiting() {
		if itr == 10 {
			stopChan <- true
		}
		itr++
	}
	assert.Equal(t, 3, itr)
}