package tests

import (
	"github.com/mostafatalebi/loadtest/pkg/stats/progress"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProgressByPercent(t *testing.T) {
	p := progress.NewProgressIndicator(99)
	p.ByPercent(99, 1, func(percent int8) {
		assert.Equal(t, int8(0), percent)
	})
	p.ByPercent(99, 10, func(percent int8) {
		assert.Equal(t, int8(10), percent)
	})
	p.ByPercent(99, 11, func(percent int8) {
		assert.Equal(t, int8(10), percent)
	})
	p.ByPercent(99, 16, func(percent int8) {
		assert.Equal(t, int8(10), percent)
	})
	p.ByPercent(99, 19, func(percent int8) {
		assert.Equal(t, int8(10), percent)
	})
	p.ByPercent(99, 99, func(percent int8) {
		assert.Equal(t, int8(100), percent)
	})
	p.ByPercent(99, 98, func(percent int8) {
		assert.Equal(t, int8(90), percent)
	})
}