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
)

type Traveler struct {
	id             int
	x              int
	y              int
	possibleMoves  [][]int
	moveChannel    chan bool
	doneSimulation chan struct{}
}

var mutex sync.Mutex

func main() {
	rand.Seed(time.Now().UnixNano())

	kratowaPlansza := make([][]int, m)
	for i := range kratowaPlansza {
		kratowaPlansza[i] = make([]int, n)
	}

	var wg sync.WaitGroup // WaitGroup do śledzenia aktywnych goroutines
	travelers := make([]*Traveler, 0)

	// Goroutine, aby generować nowych podróżników
	go func() {
		for currentID := 1; currentID <= k; currentID++ {
			x, y := rand.Intn(m), rand.Intn(n)

			// Проверяем, есть ли уже путешественник в этой клетке
			mutex.Lock()
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

			// Запоминаем путешественника в клетке
			mutex.Lock()
			kratowaPlansza[x][y] = newTraveler.id
			mutex.Unlock()

			travelers = append(travelers, newTraveler)

			wg.Add(1)
			go simulateTraveler(newTraveler, kratowaPlansza, &wg, moveChannel)

			time.Sleep(time.Second)
		}
	}()

	// Goroutine, чтобы печатать доску
	go func() {
		for {
			printKrate(kratowaPlansza, travelers)
			time.Sleep(2 * time.Second)
		}
	}()

	// Goroutine, чтобы ждать завершения всех путешественников
	go func() {
		wg.Wait()
		for _, traveler := range travelers {
			close(traveler.doneSimulation)
		}
	}()

	// Ждем завершения всех горутин
	select {}
}

func simulateTraveler(traveler *Traveler, kratowaPlansza [][]int, wg *sync.WaitGroup, moveChannel <-chan bool) {
	defer wg.Done()

	for {
		select {
		case <-time.After(2 * time.Second):

			if len(traveler.possibleMoves) > 0 {
				moveTraveler(traveler, kratowaPlansza, &mutex)
			}

		case <-moveChannel:

			traveler.possibleMoves = getEmptyNeighbors(traveler.x, traveler.y, kratowaPlansza)

		case <-traveler.doneSimulation:
			return
		}
	}
}

func moveTraveler(traveler *Traveler, kratowaPlansza [][]int, mutex *sync.Mutex) {
	// Выбираем случайное пустое место из соседних клеток
	moveIndex := rand.Intn(len(traveler.possibleMoves))
	newX, newY := traveler.possibleMoves[moveIndex][0], traveler.possibleMoves[moveIndex][1]

	mutex.Lock()
	if kratowaPlansza[newX][newY] != 0 {
		// Есть другой путешественник в этой клетке, не перемещаемся
		mutex.Unlock()
		return
	}

	// Меняем информацию о текущей клетке
	kratowaPlansza[traveler.x][traveler.y] = 0
	kratowaPlansza[newX][newY] = traveler.id
	mutex.Unlock()

	// Обновление позиции путешественника
	traveler.x = newX
	traveler.y = newY

	// Очистка possibleMoves
	traveler.possibleMoves = getEmptyNeighbors(newX, newY, kratowaPlansza)
}

func getEmptyNeighbors(x, y int, kratowaPlansza [][]int) [][]int {
	neighbors := [][]int{}

	// Влево
	if x-1 >= 0 && kratowaPlansza[x-1][y] == 0 {
		neighbors = append(neighbors, []int{x - 1, y})
	}
	// Вправо
	if x+1 < m && kratowaPlansza[x+1][y] == 0 {
		neighbors = append(neighbors, []int{x + 1, y})
	}
	// Вниз
	if y-1 >= 0 && kratowaPlansza[x][y-1] == 0 {
		neighbors = append(neighbors, []int{x, y - 1})
	}
	// Вверх
	if y+1 < n && kratowaPlansza[x][y+1] == 0 {
		neighbors = append(neighbors, []int{x, y + 1})
	}

	return neighbors
}

func printKrate(kratowaPlansza [][]int, travelers []*Traveler) {
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
				if kratowaPlansza[i][j] == 0 {
					fmt.Print(" .  ") // Пустое место
				} else {
					fmt.Printf("%02d  ", kratowaPlansza[i][j])
				}
			}

			// Отображение границ
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
