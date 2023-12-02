## <p align="center">Tango - Open Source Golang Pipeline with Backpressure (WIP)</p>

<p align="center">
  <img src="./image/logo.png" alt="">
</p>

Tango can help you implement a multi-stage pipeline that will apply backpressure on the data producer channel (will pause
reading)
in case the producer is faster than the stage processors. It may be a great choice if you want to create an ETL pipeline with
custom processing

## Concept
While building the pipeline, you must consider the case then one of the ETL stages may be slower than the rest.
E.g you may want to execute custom data enrichment with SQL queries, that means you will be reading data from the source (like Kafka)
way faster than processing them. That why we created Tango.

<p align="center">
<img src="https://miro.medium.com/v2/resize:fit:640/format:webp/1*B6KBntP9kiFBPN7NXlGruQ.gif" alt="">
</p>

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

	// Tango will consume messages from the producer channel and 
	// pass them through the stages.
	instance.SetProducerChannel(producerChannel)
	
	if err := instance.Start(); err != nil {
		panic(err)
	}
}
```

### Buffered channels

You can pass buffered channels for each stage, that will help you mitigate spikes in when a lot of updates going to the pipeline

```go
stage := tango.Stage{
  Channel: make(chan interface{}, 100),
  Function: func(msg interface{}) (interface{}, error) {
    return msg, nil
  },
},
```

### Accomplished callback
Sometimes you need to perform some extra work after the last stage is done, like messages was written to the sink
You can do this by passing callback function to  `OnProcessed` method of Tango instance

```go
tangoInstance.OnProcessed(func(i interface{}, err error) {
    if err != nil {
        fmt.Errorf("Messages happened in the last stage %v", err)
    }
    fmt.Println("Message processed", i)
})
```