package print

import (
	"fmt"
	"sync"
)

var percentHolder = make(map[int]int)
var prcLock = &sync.Mutex{}
func ProgressByPercent(total, current int64){
	prcLock.Lock()
	defer prcLock.Unlock()
	var remainder = total % 10
	var each10Percent = (total-remainder) / 10

	for i := int(1); i < 11; i++ {
		if _, ok := percentHolder[i]; ok {
			continue
		}
		if int64(i)*each10Percent == current && i != 0 && i != 10 {
			percentHolder[i] = i
			fmt.Printf("===[%v%v0]", `%`, i)
		} else if int64(i)*each10Percent == current && i == 10 {
			percentHolder[i] = i
			fmt.Printf("===[%v Completed!]", int64(i)*each10Percent)
		}
	}
}
