package suite

type CaseInfo struct {
	SuiteName     string
	MethodName    string
	TagStr        string
	Data          []string //just work on test method
	IsSkip        bool
	ParallelCount int //just work on test method
	//RunTimes      int64
	//RunDuration   int // The unit is in seconds
	//RunInterval   int // The unit is in seconds
}
