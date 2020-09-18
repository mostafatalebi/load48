package print

import (
	"fmt"
)

func ProgressByPercent(total, current int64){
	var remainder = total % 10
	var each10Percent = (total-remainder) / 10

	for i := int64(1); i < 11; i++ {
		if i*each10Percent == current && i != 0 && i != 10 {
			fmt.Printf("->-> %v%v0 ", `%`, i)
		} else if i*each10Percent == current && i == 10 {
			fmt.Printf("-> [%v Completed!]", i*each10Percent)
		}
	}
}
