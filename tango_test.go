package tango

import (
	"os"
	"testing"
	"time"
)

func TestBuildSequenceStream(t *testing.T) {
	f, err := os.Create("data.txt")
	stages := []Stage{
		{
			Channel: make(chan interface{}),
			Function: func(msg interface{}) (interface{}, error) {
				return msg, nil
			},
		},
		{
			Channel: make(chan interface{}),
			Function: func(msg interface{}) (interface{}, error) {
				if msg.(int)%5 == 0 {
					time.Sleep(time.Millisecond * 10)
				}
				return msg, nil
			},
		},
		{
			Channel: make(chan interface{}),
			Function: func(msg interface{}) (interface{}, error) {
				_, err = f.WriteString("Message")
				return msg, err
			},
		},
	}

	producerChannel := make(chan interface{})

	go func() {
		for i := 0; i <= 10000000; i++ {
			producerChannel <- i
		}
	}()

	tango := NewTango()
	tango.SetStages(stages)
	tango.SetProducerChannel(producerChannel)
	err = tango.Start()
	if err != nil {
		t.Fatalf("failed to start tango %v", err)
	}
}
