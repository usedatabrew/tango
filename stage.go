package tango

type Stage struct {
	Channel   chan interface{}
	Function  func(interface{}) (interface{}, error)
	isTheLast bool
}

type AccomplishedCallback func(interface{}, error)
