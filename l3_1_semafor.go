package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var forks = make([]sync.Mutex, 5)
var eaters = make([]sync.Mutex, 5)
var room = make(chan struct{}, 4)
var eatingListMutex sync.Mutex
var eatingList []string

func philosopher(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		think()
		room <- struct{}{} // wejście do jadalni
		pickUpForks(id)
		startEating(id)
		eat(id)
		stopEating(id)
		putDownForks(id)
		<-room // wyjście z jadalni
	}
}

func think() {
	time.Sleep(time.Millisecond * time.Duration(randInt(100, 500)))
}

func pickUpForks(id int) {
	forks[id].Lock()
	forks[(id+1)%5].Lock()
}

func eat(id int) {
	eaters[id].Lock()
	fmt.Printf("Philosopher %d is eating\n\n", id)
	time.Sleep(time.Millisecond * time.Duration(randInt(100, 500)))
	eaters[id].Unlock()
}

func putDownForks(id int) {
	forks[id].Unlock()
	forks[(id+1)%5].Unlock()
}

func startEating(id int) {
	eatingListMutex.Lock()
	eatingList = append(eatingList, fmt.Sprintf("(W%d, F%d, W%d)", id, id, (id+1)%5))
	eatingListMutex.Unlock()
	printEatingList()
}

func stopEating(id int) {
	eatingListMutex.Lock()
	for i, entry := range eatingList {
		if entry == fmt.Sprintf("(W%d, F%d, W%d)", id, id, (id+1)%5) {
			eatingList = append(eatingList[:i], eatingList[i+1:]...)
			break
		}
	}
	eatingListMutex.Unlock()
	fmt.Printf("Philosopher %d finished eating\n\n", id)
	printEatingList()
}

func printEatingList() {
	eatingListMutex.Lock()
	defer eatingListMutex.Unlock()
	fmt.Println("Lista jedzących:", eatingList)
}

func randInt(min, max int) int {
	return min + rand.Intn(max-min+1)
}

func main() {
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go philosopher(i, &wg)
	}
	wg.Wait()
}
