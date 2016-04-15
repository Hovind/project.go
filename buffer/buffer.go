package buffer

import (
	//"fmt"
	"time"

	. "project.go/obj"
)

func Manager() (chan<- Message, chan<- Message, chan<- struct{}, <-chan Message) {
	push_channel := make(chan Message)
	pop_channel := make(chan Message)
	clear_channel := make(chan struct{})


	resend_channel := make(chan Message)
	pop_success_channel := make(chan Message)

	buffer_map := make(map[int]chan<- struct{})

	go func() {
		for {
			select {
			case msg := <-push_channel:
				buffer_map[msg.Hash()] = worker(msg, resend_channel, pop_success_channel)
			case msg := <-pop_channel:
				buffer_map[msg.Hash()] <- struct{}{}
			case msg := <-pop_success_channel:
				delete(buffer_map, msg.Hash())
			case <-clear_channel:
				for _, value := range buffer_map {
					value <-struct{}{};
				}
			}
		}
	}()
	return push_channel, pop_channel, clear_channel, resend_channel
}

func worker(msg Message, resend_channel, pop_success_channel chan<- Message) chan<- struct{} {
	pop_worker_channel := make(chan struct{})
	go func() {
		for {
			select {
			case <-time.After(2 * time.Second):
				resend_channel <- msg
			case <-pop_worker_channel:
				pop_success_channel <- msg
				close(pop_worker_channel)
				return
			}
		}
	}()
	return pop_worker_channel
}
