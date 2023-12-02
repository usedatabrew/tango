package tango

import (
	"context"
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

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	for i := range t.stages {
		t.wg.Add(1)
		go func(index int) {
			defer close(t.stages[index].Channel)
			defer t.wg.Done()

			for msg := range t.stages[index].Channel {
				select {
				case <-ctx.Done():
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
							case <-ctx.Done():
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
			case <-ctx.Done():
				// Context canceled, break the loop
				log.Info("Sigterm received")
				t.wg.Done()
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

	t.wg.Wait()

	return nil
}
