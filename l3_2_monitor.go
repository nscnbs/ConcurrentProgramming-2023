package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var forks = make([]sync.Mutex, 5)
var eaters = make([]sync.Cond, 5)
var room = sync.Mutex{}
var eatingPhilosophers = struct {
	sync.Mutex
	m map[int]bool
}{m: make(map[int]bool)}

type Philosopher struct {
	id int
	sync.Mutex
}

func (p *Philosopher) think() {
	time.Sleep(time.Millisecond * time.Duration(randInt(100, 500)))
}

func (p *Philosopher) pickUpForks() {
	room.Lock()
	forks[p.id].Lock()
	forks[(p.id+1)%5].Lock()
	room.Unlock()

	p.Lock()
	eatingPhilosophers.Lock()
	eatingPhilosophers.m[p.id] = true
	eatingPhilosophers.Unlock()
	p.Unlock()
}

func (p *Philosopher) eat() {
	eaters[p.id].L.Lock()
	fmt.Printf("Philosopher %d is eating\n", p.id)
	time.Sleep(time.Millisecond * time.Duration(randInt(100, 500)))
	//fmt.Printf("Philosopher %d finished eating\n", p.id)
	eaters[p.id].L.Unlock()
	displayEatingPhilosophers()
}

func (p *Philosopher) putDownForks() {
	forks[p.id].Unlock()
	forks[(p.id+1)%5].Unlock()

	p.Lock()
	eatingPhilosophers.Lock()
	delete(eatingPhilosophers.m, p.id)
	fmt.Printf("Philosopher %d finished eating\n", p.id)
	eatingPhilosophers.Unlock()
	p.Unlock()
	displayEatingPhilosophers()

}

func displayEatingPhilosophers() {
	fmt.Printf("Lista jedzących: [")
	eatingPhilosophers.Lock()
	for id := range eatingPhilosophers.m {
		fmt.Printf("(W%d, F%d, W%d)", id, id, (id+1)%5)
	}
	fmt.Printf("]\n")
	eatingPhilosophers.Unlock()
	fmt.Println()
	//time.Sleep(time.Millisecond * 500) // Opóźnienie wyświetlania
}

func philosopher(p *Philosopher, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		p.think()
		p.pickUpForks()
		p.eat()
		p.putDownForks()
		//displayEatingPhilosophers()
	}
}

func randInt(min, max int) int {
	return min + rand.Intn(max-min+1)
}

func main() {
	var wg sync.WaitGroup
	philosophers := make([]Philosopher, 5)
	//displayEatingPhilosophers()
	for i := 0; i < 5; i++ {
		philosophers[i].id = i
		eaters[i] = *sync.NewCond(&philosophers[i])
		wg.Add(1)
		go philosopher(&philosophers[i], &wg)
	}
	wg.Wait()
}
