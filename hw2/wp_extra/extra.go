package main

import (
	"fmt"
	"strconv"
	"time"
)

const DEFAULT_POOL_SIZE int = 10

func worker(in, out chan string, cancelCh chan struct{}) {
	defer close(out)
loop:
	for {
		select {
		case data := <-in:
			//fmt.Printf("Processed: %s\n", data)
			out <- data
		case <-cancelCh:
			fmt.Println("Le kill")
			break loop
		}
	}

}

func initPool(size int) []chan struct{} {
	if size == 0 {
		size = DEFAULT_POOL_SIZE
	}

	cancelQueue := make([]chan struct{}, size)

	for i := 0; i < size; i++ {
		in, out, c := createWorker()

		cancelQueue = append(cancelQueue, c)
		go generate(in, i)
		go print(out, i)
		go worker(in, out, c)
	}

	return cancelQueue
}

func generate(in chan<- string, n int) {
	defer close(in)
	for i := n; i < 10000; i++ {
		in <- strconv.Itoa(i)
		//fmt.Printf("Generated[%v]: %v\n", i, i)
		time.Sleep(time.Second * 5)
	}
}

func print(out chan string, n int) {
	//defer close(out)
	for range out {
		//fmt.Printf("Received[%v]: %s\n", n, data)
		time.Sleep(time.Second)
	}
}

func createWorker() (chan string, chan string, chan struct{}) {
	in := make(chan string, 1)
	out := make(chan string, 1)
	c := make(chan struct{})

	return in, out, c
}

func removeWorker(q *[]chan struct{}, c chan struct{}) *[]chan struct{} {
	c <- struct{}{}

	tmp := make([]chan struct{}, 0)

	tmp = append(tmp, (*q)[:len(*q)-1]...)

	tmp = append(tmp, (*q)[len(*q):]...)

	return &tmp
}

func listen(q *[]chan struct{}, cmdCh chan string) {
	for cmd := range cmdCh {
		switch cmd {
		case "add":
			in, out, c := createWorker()

			*q = append(*q, c)
			printQlen(q)
			fmt.Println("add")
			go generate(in, len(*q))
			go print(out, len(*q))
			go worker(in, out, c)
			printQlen(q)

		case "remove":
			printQlen(q)
			fmt.Println("remove")

			c := (*q)[len(*q)-1]

			q = removeWorker(q, c)
			printQlen(q)
		default:
			fmt.Printf("Unknown command: %v", cmd)
			continue
		}
	}
}

func printQlen(q *[]chan struct{}) {
	fmt.Printf("Workers running: %v\n", len(*q))
	time.Sleep(time.Second)
}

func main() {
	fmt.Println("Starting pool")
	q := initPool(0)
	cmdCh := make(chan string)
	go listen(&q, cmdCh)

	cmdCh <- "add"
	time.Sleep(time.Second)
	cmdCh <- "add"
	time.Sleep(time.Second)
	cmdCh <- "add"
	time.Sleep(time.Second)
	cmdCh <- "add"
	time.Sleep(time.Second)
	cmdCh <- "add"
	time.Sleep(time.Second)
	cmdCh <- "remove"
	time.Sleep(time.Second)
	cmdCh <- "remove"
	time.Sleep(time.Second)
	cmdCh <- "remove"
	time.Sleep(time.Second)
	cmdCh <- "remove"
	time.Sleep(time.Second)
	cmdCh <- "remove"
}
