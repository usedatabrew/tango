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

import "github.com/usedatabrew/tango"

func main() {
	instance := tango.NewTango()
	stage := tango.Stage{
	  Channel: make(chan interface{}),
	  Function: func(message interface{}, err error) (interface{}, error) {
		  // @todo do your processing for the message
		  return message, err
	  }, 
	}
	instance.AppendStage(stage)

	if err := instance.Start(); err != nil {
		panic(err)
	}
}
```