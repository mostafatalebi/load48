package stats

var DefaultPresetWithAutoFailedCodes = map[string]string{
	TotalSent:            "Total Sent Number of Requests",
	Success:              "Total Success",
	Timeout:              "Total Timeouts",
	ConnRefused:          "Total Conn. Refused",
	AverageExecDuration:  "Average App Execution",
	AverageDuration:      "Average Duration",
	ShortestDuration:     "Shortest Duration",
	ShortestExecDuration: "Shortest App Execution",
	LongestDuration:      "Longest Duration",
	LongestExecDuration:  "Longest App Execution",
	MaxConcurrencyAchieved:  "Max Concurrency Achieved",
}
