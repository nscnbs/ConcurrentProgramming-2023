package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const (
	m = 10      // Liczba rzędów
	n = 10      // Liczba kolumn
	k = m*n - 1 // Maksymalna liczba podróżników
	d = 5       // Ograniczony czas życia dzikiego lokatora
	z = 3       // Ograniczony czas istnienia czasowego zagrożenia
)

type Traveler struct {
	id             int
	x              int
	y              int
	possibleMoves  [][]int
	moveChannel    chan bool
	doneSimulation chan struct{}
	travelerMutex  sync.Mutex
}

type WildLodger struct {
	x     int
	y     int
	life  int
	mutex sync.Mutex
}

type TemporaryThreat struct {
	x     int
	y     int
	life  int
	mutex sync.Mutex
}

var mutex sync.Mutex

func main() {
	rand.Seed(time.Now().UnixNano())

	kratowaPlansza := make([][]int, m)
	for i := range kratowaPlansza {
		kratowaPlansza[i] = make([]int, n)
	}

	var wg sync.WaitGroup
	travelers := make([]*Traveler, 0)
	wildLodgers := make([]*WildLodger, 0)
	temporaryThreats := make([]*TemporaryThreat, 0)

	// Goroutine do generowania podróżników
	go func() {
		for currentID := 1; currentID <= k; currentID++ {
			mutex.Lock()
			x, y := rand.Intn(m), rand.Intn(n)
			for kratowaPlansza[x][y] != 0 {
				x, y = rand.Intn(m), rand.Intn(n)
			}
			mutex.Unlock()

			moveChannel := make(chan bool)
			doneSimulation := make(chan struct{})

			newTraveler := &Traveler{
				id:             currentID,
				x:              x,
				y:              y,
				possibleMoves:  getEmptyNeighbors(x, y, kratowaPlansza),
				moveChannel:    moveChannel,
				doneSimulation: doneSimulation,
			}

			mutex.Lock()
			kratowaPlansza[x][y] = newTraveler.id
			mutex.Unlock()

			travelers = append(travelers, newTraveler)

			wg.Add(1)
			go simulateTraveler(newTraveler, kratowaPlansza, &wg, moveChannel, wildLodgers, temporaryThreats)

			time.Sleep(time.Second)
		}
	}()

	// Goroutine do generowania dzikich lokatorów
	go func() {
		for {
			x, y := rand.Intn(m), rand.Intn(n)

			mutex.Lock()
			// Sprawdzamy, czy już istnieje dziki lokator w tej komórce
			wildLodgerExists := false
			for _, wildLodger := range wildLodgers {
				if wildLodger.x == x && wildLodger.y == y {
					wildLodgerExists = true
					break
				}
			}
			// Sprawdzamy, czy komórka jest wolna
			if kratowaPlansza[x][y] == 0 && !wildLodgerExists {
				newWildLodger := &WildLodger{
					x:    x,
					y:    y,
					life: d,
				}
				wildLodgers = append(wildLodgers, newWildLodger)
				kratowaPlansza[x][y] = -1 // Oznaczamy dzikiego lokatora inaczej niż podróżników
			}
			mutex.Unlock()

			time.Sleep(3 * time.Second)
		}
	}()

	// Goroutine do generowania czasowych zagrożeń
	go func() {
		for {
			x, y := rand.Intn(m), rand.Intn(n)

			mutex.Lock()
			// Sprawdzamy, czy już istnieje czasowe zagrożenie w tej komórce
			temporaryThreatExists := false
			for _, temporaryThreat := range temporaryThreats {
				if temporaryThreat.x == x && temporaryThreat.y == y {
					temporaryThreatExists = true
					break
				}
			}
			// Sprawdzamy, czy komórka jest wolna
			if kratowaPlansza[x][y] == 0 && !temporaryThreatExists {
				newTemporaryThreat := &TemporaryThreat{
					x:    x,
					y:    y,
					life: z,
				}
				temporaryThreats = append(temporaryThreats, newTemporaryThreat)
				kratowaPlansza[x][y] = -2 // Oznaczamy czasowe zagrożenie inaczej niż podróżników
			}
			mutex.Unlock()

			time.Sleep(5 * time.Second)
		}
	}()

	// Goroutine do wyświetlania planszy
	go func() {
		for {
			updateLife(wildLodgers, temporaryThreats, kratowaPlansza, &mutex)
			printKrate(kratowaPlansza, travelers, wildLodgers, temporaryThreats, &mutex)
			time.Sleep(2 * time.Second)
		}
	}()

	// Goroutine do oczekiwania na zakończenie symulacji
	go func() {
		wg.Wait()
		for _, traveler := range travelers {
			close(traveler.doneSimulation)
		}
	}()

	// Czekamy na zakończenie programu
	select {}
}

func simulateTraveler(traveler *Traveler, kratowaPlansza [][]int, wg *sync.WaitGroup, moveChannel <-chan bool, wildLodgers []*WildLodger, temporaryThreats []*TemporaryThreat) {
	defer wg.Done()

	for {

		select {
		case <-time.After(2 * time.Second):

			if len(traveler.possibleMoves) > 0 {
				moveTraveler(traveler, kratowaPlansza, &mutex, wildLodgers, temporaryThreats)
			}

			// Sprawdzamy, czy w tej komórce jest dziki lokator

			for _, wildLodger := range wildLodgers {
				if wildLodger.x == traveler.x && wildLodger.y == traveler.y {
					// Jeśli tak, podróżnik musi opuścić tę komórkę
					moveTraveler(traveler, kratowaPlansza, &mutex, wildLodgers, temporaryThreats)
				}
			}

			// Sprawdzamy, czy w tej komórce jest czasowe zagrożenie

			for _, temporaryThreat := range temporaryThreats {
				if temporaryThreat.x == traveler.x && temporaryThreat.y == traveler.y {
					mutex.Lock()
					// Jeśli tak, podróżnik przestaje istnieć
					kratowaPlansza[traveler.x][traveler.y] = 0
					mutex.Unlock()
					return
				}
			}

		case <-moveChannel:
			traveler.possibleMoves = getEmptyNeighbors(traveler.x, traveler.y, kratowaPlansza)
		case <-traveler.doneSimulation:

			return
		}

	}
}

func moveTraveler(traveler *Traveler, kratowaPlansza [][]int, mutex *sync.Mutex, wildLodgers []*WildLodger, temporaryThreats []*TemporaryThreat) {
	// Wybieramy losową pustą komórkę spośród sąsiadujących komórek
	moveIndex := rand.Intn(len(traveler.possibleMoves))
	newX, newY := traveler.possibleMoves[moveIndex][0], traveler.possibleMoves[moveIndex][1]

	mutex.Lock()
	// Jeśli ruch na miejsce dzikiego lokatora
	if kratowaPlansza[newX][newY] == -1 {
		moveWildLodgerToEmpty(wildLodgers, newX, newY, kratowaPlansza, mutex)
	} else if kratowaPlansza[newX][newY] == -2 { // Jeśli ruch na miejsce czasowego zagrożenia
		removeTemporaryThreat(newX, newY, temporaryThreats, kratowaPlansza, mutex)
	} else if kratowaPlansza[newX][newY] != 0 {
		// Inny podróżnik jest już w tej komórce, nie przemieszczamy się
		mutex.Unlock()
		return
	}

	// Zmieniamy informacje o obecnej komórce
	kratowaPlansza[traveler.x][traveler.y] = 0
	kratowaPlansza[newX][newY] = traveler.id
	mutex.Unlock()

	// Aktualizacja pozycji podróżnika
	traveler.x = newX
	traveler.y = newY

	// Czyszczenie possibleMoves
	traveler.possibleMoves = getEmptyNeighbors(newX, newY, kratowaPlansza)
}

func moveWildLodgerToEmpty(wildLodgers []*WildLodger, x, y int, kratowaPlansza [][]int, mutex *sync.Mutex) {
	for _, wildLodger := range wildLodgers {
		if wildLodger.x == x && wildLodger.y == y {
			moveWildLodger(wildLodger, kratowaPlansza, mutex)
			break
		}
	}
	// Убедитесь, что изменения отражаются на оригинальном слайсе
	for i, wildLodger := range wildLodgers {
		if wildLodger.x == x && wildLodger.y == y {
			wildLodgers[i] = wildLodger
			break
		}
	}
}

func removeTemporaryThreat(x, y int, temporaryThreats []*TemporaryThreat, kratowaPlansza [][]int, mutex *sync.Mutex) {

	for i, temporaryThreat := range temporaryThreats {
		if temporaryThreat.x == x && temporaryThreat.y == y {
			if kratowaPlansza[x][y] == -2 {
				kratowaPlansza[x][y] = 0
			}

			mutex.Lock()
			temporaryThreats = append(temporaryThreats[:i], temporaryThreats[i+1:]...)
			mutex.Unlock()
			break
		}
	}
}

func updateLife(wildLodgers []*WildLodger, temporaryThreats []*TemporaryThreat, kratowaPlansza [][]int, mutex *sync.Mutex) {
	if len(wildLodgers) == 0 && len(temporaryThreats) == 0 {
		return
	}
	wildLodgersCopy := make([]*WildLodger, len(wildLodgers))
	temporaryThreatsCopy := make([]*TemporaryThreat, len(temporaryThreats))
	copy(wildLodgersCopy, wildLodgers)
	copy(temporaryThreatsCopy, temporaryThreats)

	for _, wildLodger := range wildLodgersCopy {
		wildLodger.life--
		if wildLodger.life <= 0 {
			mutex.Lock()
			if kratowaPlansza[wildLodger.x][wildLodger.y] == -1 {
				kratowaPlansza[wildLodger.x][wildLodger.y] = 0
			}
			mutex.Unlock()
		}
	}

	for _, temporaryThreat := range temporaryThreatsCopy {
		temporaryThreat.life--
		if temporaryThreat.life <= 0 {
			mutex.Lock()
			if kratowaPlansza[temporaryThreat.x][temporaryThreat.y] == -2 {
				kratowaPlansza[temporaryThreat.x][temporaryThreat.y] = 0
			}
			mutex.Unlock()
		}
	}
}

func moveWildLodger(wildLodger *WildLodger, kratowaPlansza [][]int, mutex *sync.Mutex) {
	nearestEmpty := findNearestEmpty(wildLodger.x, wildLodger.y, kratowaPlansza)
	if nearestEmpty != nil {
		newX, newY := nearestEmpty[0], nearestEmpty[1]

		mutex.Lock()
		if kratowaPlansza[wildLodger.x][wildLodger.y] == -1 {
			kratowaPlansza[wildLodger.x][wildLodger.y] = 0
			kratowaPlansza[newX][newY] = -1
		}
		mutex.Unlock()

		wildLodger.x = newX
		wildLodger.y = newY
	}
}

func findNearestEmpty(x, y int, kratowaPlansza [][]int) []int {
	// Kolejność: lewo, prawo, góra, dół
	directions := [][]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}

	for _, dir := range directions {
		newX, newY := x+dir[0], y+dir[1]
		if newX >= 0 && newX < m && newY >= 0 && newY < n && kratowaPlansza[newX][newY] == 0 {
			return []int{newX, newY}
		}
	}

	return nil
}

func getEmptyNeighbors(x, y int, kratowaPlansza [][]int) [][]int {
	neighbors := [][]int{}

	// W lewo
	if x-1 >= 0 && kratowaPlansza[x-1][y] == 0 {
		neighbors = append(neighbors, []int{x - 1, y})
	}
	// W prawo
	if x+1 < m && kratowaPlansza[x+1][y] == 0 {
		neighbors = append(neighbors, []int{x + 1, y})
	}
	// W górę
	if y-1 >= 0 && kratowaPlansza[x][y-1] == 0 {
		neighbors = append(neighbors, []int{x, y - 1})
	}
	// W dół
	if y+1 < n && kratowaPlansza[x][y+1] == 0 {
		neighbors = append(neighbors, []int{x, y + 1})
	}

	return neighbors
}

func printKrate(kratowaPlansza [][]int, travelers []*Traveler, wildLodgers []*WildLodger, temporaryThreats []*TemporaryThreat, mutex *sync.Mutex) {
	fmt.Println("Stan planszy:")

	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			foundTraveler := false
			for _, traveler := range travelers {
				if traveler.x == i && traveler.y == j {
					fmt.Printf("%02d  ", traveler.id)
					foundTraveler = true
					break
				}
			}
			if !foundTraveler {
				foundWildLodger := false
				for _, wildLodger := range wildLodgers {
					if wildLodger.x == i && wildLodger.y == j {
						fmt.Print(" *  ") // Dziki lokator
						foundWildLodger = true
						break
					}
				}
				if !foundWildLodger {
					foundTemporaryThreat := false
					for _, temporaryThreat := range temporaryThreats {
						if temporaryThreat.x == i && temporaryThreat.y == j {
							fmt.Print(" #  ") // Czasowe zagrożenie
							foundTemporaryThreat = true
							break
						}
					}
					if !foundTemporaryThreat {
						if kratowaPlansza[i][j] == 0 {
							fmt.Print("    ") // Puste miejsce
						} else {
							fmt.Printf("%02d  ", kratowaPlansza[i][j])
						}
					}
				}
			}

			// Wyświetlanie krawędzi
			if j < n-1 {
				fmt.Print("|")
			}
		}
		fmt.Println()
		if i < m-1 {
			for j := 0; j < n; j++ {
				fmt.Print("----")
				if j < n-1 {
					fmt.Print("+")
				}
			}
			fmt.Println()
		}
	}
}
