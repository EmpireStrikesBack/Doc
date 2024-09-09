package main

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"runtime"
	"sync"
	"text/tabwriter"
	"time"
)

var wg sync.WaitGroup

func main() {

	// Creating a join point
	for _, salutation := range []string{"hello", "greetings", "good day"} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Println(salutation)
		}()
	}
	wg.Wait()

	memConsumed := func() uint64 {
		runtime.GC()
		var s runtime.MemStats
		runtime.ReadMemStats(&s)
		return s.Sys
	}

	// Measuring the amount of memory allocated before-after goroutine creation
	var c <-chan interface{}
	noop := func() { wg.Done(); <-c }

	const numGoroutines = 1e4
	wg.Add(numGoroutines)
	before := memConsumed()
	for i := numGoroutines; i > 0; i-- {
		go noop()
	}
	wg.Wait()
	after := memConsumed()
	fmt.Printf("%.3fkb\n", float64(after-before)/numGoroutines/1000)

	// WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("1st goroutine sleeping...")
		time.Sleep(1)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("2cd goroutine sleeping...")
		time.Sleep(2)
	}()
	wg.Wait()
	fmt.Println("all goroutines complete.")

	// Couple calls to Add as closely as possible to the goroutines
	hello := func(wg *sync.WaitGroup, id int) {
		defer wg.Done()
		fmt.Printf("Hello from %v\n", id)
	}
	const numGreeeters = 5
	wg.Add(numGreeeters)
	for i := 0; i < numGreeeters; i++ {
		go hello(&wg, i+1)
	}
	wg.Wait()

	// Mutex : 2 goroutines attempting to increment & decrement a common value
	var count int
	var lock sync.Mutex
	var arithmetic sync.WaitGroup

	increment := func() {
		lock.Lock()
		defer lock.Unlock()
		count++
		fmt.Printf("Incrementing: %d\n", count)
	}

	decrement := func() {
		lock.Lock()
		defer lock.Unlock()
		count--
		fmt.Printf("Decrementing: %d\n", count)
	}

	for i := 0; i <= 5; i++ {
		arithmetic.Add(1)
		go func() {
			defer arithmetic.Done()
			decrement()
		}()
		arithmetic.Add(1)
		go func() {
			defer arithmetic.Done()
			increment()
		}()
	}
	arithmetic.Wait()
	fmt.Println("Arithmetic complete.")

	// WRMutex
	producer := func(wg *sync.WaitGroup, l sync.Locker) {
		defer wg.Done()
		for i := 5; i > 0; i-- {
			l.Lock()
			l.Unlock()
			time.Sleep(1)
		}
	}

	observer := func(wg *sync.WaitGroup, l sync.Locker) {
		defer wg.Done()
		l.Lock()
		defer l.Unlock()
	}

	test := func(count int, mutex, rwMutex sync.Locker) time.Duration {
		wg.Add(count + 1)
		beginTestTime := time.Now()
		go producer(&wg, mutex)
		for i := count; i > 0; i-- {
			go observer(&wg, rwMutex)
		}
		wg.Wait()
		return time.Since(beginTestTime)
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 1, 2, ' ', 0)
	defer tw.Flush()

	var m sync.RWMutex
	fmt.Fprintf(tw, "Readers\tRWMutex\tMutex\n")
	for i := 0; i < 20; i++ {
		count := int(math.Pow(2, float64(i)))
		fmt.Fprintf(
			tw,
			"%d\t%v\t%v\n",
			count,
			test(count, &m, m.RLocker()),
			test(count, &m, &m),
		)
	}

	// Cond : rendez-vous point for goroutines waiting for or announcing the occurence of an event \\

	// Signal
	a := sync.NewCond(&sync.Mutex{})
	queue := make([]interface{}, 0, 10)

	removeFromQueue := func(delay time.Duration) {
		time.Sleep(delay)
		a.L.Lock()
		queue = queue[1:]
		fmt.Println("Removed from queue")
		a.L.Unlock()
		a.Signal()
	}

	for i := 0; i < 10; i++ {
		a.L.Lock()
		for len(queue) == 2 {
			a.Wait()
		}
		fmt.Println("Adding to queue")
		queue = append(queue, struct{}{})
		go removeFromQueue(1 * time.Second)
		a.L.Unlock()
	}

	// Broadcast
	type Button struct {
		Clicked *sync.Cond
	}
	button := Button{Clicked: sync.NewCond(&sync.Mutex{})}

	var goroutineRunning sync.WaitGroup
	subscribe := func(d *sync.Cond, fn func()) {
		goroutineRunning.Add(1)
		go func() {
			goroutineRunning.Done()
			d.L.Lock()
			defer d.L.Unlock()
			d.Wait()
			fn()
		}()
		goroutineRunning.Wait()
	}

	var clickRegistered sync.WaitGroup
	clickRegistered.Add(3)
	subscribe(button.Clicked, func() {
		fmt.Println("Maximizing window.")
		clickRegistered.Done()
	})
	subscribe(button.Clicked, func() {
		fmt.Println("Displaying annoying dialog box!")
		clickRegistered.Done()
	})
	subscribe(button.Clicked, func() {
		fmt.Println("Mouse cliked.")
		clickRegistered.Done()
	})
	button.Clicked.Broadcast()
	clickRegistered.Wait()

	// Once
	increment1 := func() {
		count++
	}

	var once sync.Once
	var increments sync.WaitGroup
	increments.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer increments.Done()
			once.Do(increment1)
		}()
	}
	increments.Wait()
	fmt.Printf("count is %d\n", count)

	//Pool
	myPool := &sync.Pool{
		New: func() interface{} {
			fmt.Println("Creating new instance.")
			return struct{}{}
		},
	}

	myPool.Get()
	instance := myPool.Get()
	myPool.Put(instance)
	myPool.Get()

	var numcCalcsCreated int
	calcPool := &sync.Pool{
		New: func() interface{} {
			numcCalcsCreated += 1
			mem := make([]byte, 1024)
			return &mem
		},
	}

	calcPool.Put(calcPool.New())
	calcPool.Put(calcPool.New())
	calcPool.Put(calcPool.New())
	calcPool.Put(calcPool.New())

	const numWorkers = 1024 * 1024
	wg.Add(numWorkers)
	for i := numWorkers; i > 0; i-- {
		go func() {
			defer wg.Done()
			mem := calcPool.Get().(*[]byte)
			defer calcPool.Put(mem)
		}()
	}
	wg.Wait()
	fmt.Printf("%d calculators were created.\n", numcCalcsCreated)

	// Channels \\

	// String channel
	stringStream := make(chan string)
	go func() {
		stringStream <- "Hello channels"
	}()
	fmt.Println(<-stringStream)

	// receiving with returning 2 values
	go func() {
		stringStream <- "Hello my channels"
	}()
	salutation, ok := <-stringStream
	fmt.Printf("(%v): %v\n", ok, salutation)

	// Reading from a closed channel
	intStream := make(chan int)
	close(intStream)
	integer, ok := <-intStream
	fmt.Printf("(%v): %v\n", ok, integer)

	//Ranging over channels
	intStream1 := make(chan int)
	go func() {
		defer close(intStream1)
		for i := 1; i <= 5; i++ {
			intStream1 <- i
		}
	}()

	for integer := range intStream1 {
		fmt.Printf("%v", integer)
	}
	fmt.Print("\n")

	// Closing channel is both cheaper and faster than performing n writes
	begin := make(chan interface{})
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-begin
			fmt.Printf("%v has begun\n", i)
		}(i)
	}
	fmt.Println("Unblocking goroutines")
	close(begin)
	wg.Wait()

	// Another more complete example using buffered channels
	var stdoutBuff bytes.Buffer
	defer stdoutBuff.WriteTo(os.Stdout)

	intStream2 := make(chan int, 4)
	go func() {
		defer close(intStream2)
		defer fmt.Fprintln(&stdoutBuff, "Producer done.")
		for i := 0; i < 5; i++ {
			fmt.Fprintf(&stdoutBuff, "sending: %d\n", i)
			intStream2 <- i
		}
	}()

	for integer := range intStream2 {
		fmt.Fprintf(&stdoutBuff, "Received %v.\n", integer)
	}

	// Goroutine that owns a channel and a consumer that handles blocking & closing of a channel
	chanOwner := func() <-chan int {
		resultStream := make(chan int, 5)
		go func() {
			defer close(resultStream)
			for i := 0; i < 6; i++ {
				resultStream <- i
			}
		}()
		return resultStream
	}

	resultStream := chanOwner()
	for result := range resultStream {
		fmt.Printf("Received: %d\n", result)
	}
	fmt.Println("Done receiving !!")

	// Select statements \\

	// first example
	start := time.Now()
	b := make(chan interface{})
	go func() {
		time.Sleep(5 * time.Second)
		close(b)
	}()

	fmt.Println("Blocking on read...")
	select {
	case <-b:
		fmt.Printf("Unblocked %v later.\n", time.Since(start))
	}

	// What happens if multiple channels are ready simultaneously
	c1 := make(chan interface{})
	close(c1)
	c2 := make(chan interface{})
	close(c2)

	var c1Count, c2Count int
	for i := 1000; i >= 0; i-- {
		select {
		case <-c1:
			c1Count++
		case <-c2:
			c2Count++
		}
	}
	fmt.Printf("c1Count: %d\nc2Count: %d\n", c1Count, c2Count)

	// What happens if there are never any channel that become ready
	var z <-chan int
	select {
	case <-z:
	case <-time.After(1 * time.Second):
		fmt.Println("Timed out.")
	}

	// What happens when no channel us ready & we need to do something in the meantime
	start1 := time.Now()
	var c3, c4 <-chan int
	select {
	case <-c3:
	case <-c4:
	default:
		fmt.Printf("In default after %v\n\n", time.Since(start1))
	}

	// using the default clause in conjunction with a for-select loop
	done := make(chan interface{})
	go func() {
		time.Sleep(5 * time.Second)
		close(done)
	}()

	workCounter := 0
loop:
	for {
		select {
		case <-done:
			break loop
		default:
			workCounter++
			time.Sleep(1 * time.Second)
		}
	}
	fmt.Printf("Achieved %v cycle of work before signalled to stop.\n", workCounter)
}
