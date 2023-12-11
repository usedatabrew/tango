package tango

import (
	"errors"
	"github.com/charmbracelet/log"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

type Tango struct {
	stages          []Stage
	wg              sync.WaitGroup
	producerChannel chan interface{}
	osSignChannel   chan os.Signal
	onAccomplished  AccomplishedCallback
	closed          atomic.Bool
}

func NewTango() *Tango {
	tango := &Tango{
		osSignChannel:   make(chan os.Signal, 1),
		wg:              sync.WaitGroup{},
		producerChannel: make(chan interface{}),
	}
	tango.closed.Store(false)

	return tango
}

func (t *Tango) SetProducerChannel(producerChannel chan interface{}) {
	t.producerChannel = producerChannel
}

func (t *Tango) SetStages(stages []Stage) {
	if len(stages) == 0 {
		panic("cant set empty stages")
	}

	stages[len(stages)-1].isTheLast = true
	t.stages = stages
}

func (t *Tango) AppendStage(stage Stage) {
	if len(t.stages) != 0 {
		// reset isTheLast flag for exising stages
		t.stages[len(t.stages)-1].isTheLast = false
	}

	stage.isTheLast = true
	t.stages = append(t.stages, stage)
}

func (t *Tango) OnProcessed(callback AccomplishedCallback) {
	t.onAccomplished = callback
}

func (t *Tango) Start() error {
	if t.producerChannel == nil {
		return errors.New("producer channel is nil. Cant start stages without producer")
	}

	if len(t.stages) == 0 {
		return errors.New("no stages found. cant start without stages")
	}

	done := make(chan struct{})
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	for i := range t.stages {
		go func(index int) {
			defer func() {
				close(t.stages[index].Channel)
			}()

			for msg := range t.stages[index].Channel {
				select {
				case <-done:
					t.closed.Store(true)
					break
				default:
					result, err := t.stages[index].Function(msg)
					if err == nil {
						if t.stages[index].isTheLast {
							if t.onAccomplished != nil {
								t.onAccomplished(result, err)
							}

						}
						if index+1 < len(t.stages) {
							select {
							case t.stages[index+1].Channel <- result:
								// Message sent successfully to the next stage
							case <-done:
								// Context canceled, break the loop
								break
							}
						}
					} else {
						if t.onAccomplished != nil {
							t.onAccomplished(nil, errors.New("failed to process message"))
						}

						log.Fatalf("Error in Stage %d: %v\n", index+1, err)
						close(t.stages[0].Channel)
						return
					}
				}
			}
		}(i)
	}

	t.wg.Add(1)
	go func() {
		for {
			select {
			case <-done:
				if t.closed.Load() {
					break
				}
				t.closed.Store(true)
				close(done)
				break
			case msg, ok := <-t.producerChannel:
				if !ok {
					// producerChannel closed, break the loop
					return
				}

				select {
				case t.stages[0].Channel <- msg:
					// Message sent successfully to the first stage
				case <-time.After(time.Second):
					// Backpressure: First stage is not ready to receive, wait for a short period

					if t.closed.Load() {
						break
					}

					t.stages[0].Channel <- msg
				}
			}
		}
	}()

	<-sigCh

	return nil
}
