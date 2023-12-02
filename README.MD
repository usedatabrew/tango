## <p align="center">Tango - Open Source Golang Pipeline with Backpressure (WIP)</p>

<p align="center">
  <img src="./image/logo.png" alt="">
</p>

Tango can help you implement a multi-stage pipeline that will apply backpressure on the data producer channel (will pause
reading)
in case the producer is faster than the stage processors. It may be a great choice if you want to create an ETL pipeline with
custom processing


## Usage example

```go
package main

import (
	"github.com/usedatabrew/tango"
	"time"
)

func main() {
	instance := tango.NewTango()

	stages := []tango.Stage{
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
				return msg, nil
			},
		},
	}
	
	instance.SetStages(stages)

	producerChannel := make(chan interface{})

	go func() {
		for i := 0; i <= 10000000; i++ {
			producerChannel <- i
			time.Sleep(time.Millisecond * 200)
		}
	}()

	if err := instance.Start(); err != nil {
		panic(err)
	}
}
```