package main

import (
	"fmt"
	"math/rand"
	"sync"

	"github.com/fatih/color"
)

const NumberOfPizzas = 10

var pizzasMade, pizzasFailed, total int

// Ingredient represents an ingredient with its name and quantity.
type Ingredient struct {
	Name     string
	Quantity int
}

// PizzaType represents a pizza type with its name and required ingredients.
type PizzaType struct {
	Name        string
	Ingredients []Ingredient
}

// AvailableIngredients represents the stock of available ingredients.
var AvailableIngredients = map[string]int{
	"Tomato sauce":  10,
	"Mozzarella":    10,
	"Bacon":         10,
	"Eggs":          9,
	"Onions":        8,
	"Chili peppers": 8,
	"Anchovies":     5,
	"Olive oil":     3,
}

// Producer is a type for structs that holds two channels: one for pizzas, with all
// information for a given pizza order including whether it was made
// successfully, and another to handle end of processing (when we quit the channel)
type Producer struct {
	data            chan PizzaOrder
	quit            chan chan error
	wg              *sync.WaitGroup
	ingredientMutex sync.Mutex
}

// PizzaOrder is a type for structs that describes a given pizza order. It has the order
// number, a message indicating what happened to the order, and a boolean
// indicating if the order was successfully completed.
type PizzaOrder struct {
	pizzaNumber int
	message     string
	success     bool
}

// Close is simply a method of closing the channel when we are done with it (i.e.
// something is pushed to the quit channel)
func (p *Producer) Close() error {
	ch := make(chan error)
	p.quit <- ch
	return <-ch
}

// makePizza attempts to make a pizza. If the required ingredients for the pizza type
// are available, it decrements the ingredient quantities and marks the pizza as successful.
// Otherwise, it marks the pizza as failed.
func makePizza(pizzaNumber int, pizzaType PizzaType) *PizzaOrder {
	pizzaNumber++
	if pizzaNumber <= NumberOfPizzas {
		msg := ""
		success := false
		total++

		pizzaIngredients := pizzaType.Ingredients
		for _, ingredient := range pizzaIngredients {
			if AvailableIngredients[ingredient.Name] < ingredient.Quantity {
				pizzasFailed++
				msg = fmt.Sprintf("*** We ran out of %s for pizza #%d!", ingredient.Name, pizzaNumber)
				p := PizzaOrder{
					pizzaNumber: pizzaNumber,
					message:     msg,
					success:     success,
				}
				return &p
			}
		}

		for _, ingredient := range pizzaIngredients {
			decrementIngredient(ingredient.Name, ingredient.Quantity)
		}

		pizzasMade++

		msgDelay := fmt.Sprintf("Received order #%d! for %s", pizzaNumber, pizzaType.Name)
		color.Cyan(msgDelay)

		success = true
		msg = fmt.Sprintf("Pizza order #%d is ready!", pizzaNumber)

		p := PizzaOrder{
			pizzaNumber: pizzaNumber,
			message:     msg,
			success:     success,
		}

		return &p
	}

	return &PizzaOrder{
		pizzaNumber: pizzaNumber,
	}
}

// decrementIngredient decrements the quantity of a given ingredient in the available stock.
func decrementIngredient(ingredientName string, quantity int) {
	AvailableIngredients[ingredientName] -= quantity
}

// pizzeria is a goroutine that```go
// pizzeria is a goroutine that runs in the background and
// calls makePizza to try to make one order each time it iterates through
// the for loop. It executes until it receives something on the quit
// channel. The quit channel does not receive anything until the consumer
// sends it (when the number of orders is greater than or equal to the
// constant NumberOfPizzas).
func pizzeria(pizzaMaker *Producer, pizzaTypes []PizzaType) {
	// keep track of which pizza we are making
	var i = 0

	// this loop will continue to execute, trying to make pizzas,
	// until the quit channel receives something.
	for {
		pizzaTypeIndex := rand.Intn(len(pizzaTypes))
		pizzaType := pizzaTypes[pizzaTypeIndex]

		currentPizza := makePizza(i, pizzaType)
		if currentPizza != nil {
			i = currentPizza.pizzaNumber
			select {
			// we tried to make a pizza (we send something to the data channel -- a chan PizzaOrder)
			case pizzaMaker.data <- *currentPizza:
			// we want to quit, so send pizzaMaker.quit to the quitChan (a chan error)
			case quitChan := <-pizzaMaker.quit:
				// close channels
				close(pizzaMaker.data)
				close(quitChan)
				pizzaMaker.wg.Done() // Signal that the pizzeria goroutine has finished
				return
			}
		}
	}
}

func main() {
	// print out a message
	color.Cyan("The Pizzeria is open for business!")
	color.Cyan("----------------------------------")

	// Define pizza types and their required ingredients
	pizzaTypes := []PizzaType{
		{
			Name: "Mazza",
			Ingredients: []Ingredient{
				{Name: "Tomato sauce", Quantity: 1},
				{Name: "Mozzarella", Quantity: 1},
				{Name: "Bacon", Quantity: 1},
				{Name: "Eggs", Quantity: 1},
				{Name: "Onions", Quantity: 1},
				{Name: "Chili peppers", Quantity: 1},
			},
		},
		{
			Name: "Napoletana",
			Ingredients: []Ingredient{
				{Name: "Tomato sauce", Quantity: 1},
				{Name: "Mozzarella", Quantity: 1},
				{Name: "Anchovies", Quantity: 1},
				{Name: "Olive oil", Quantity: 1},
			},
		},
	}

	// create a producer
	pizzaJob := &Producer{
		data: make(chan PizzaOrder),
		quit: make(chan chan error),
		wg:   &sync.WaitGroup{},
	}

	// Add 1 to the WaitGroup to indicate the pizzeria goroutine
	pizzaJob.wg.Add(1)

	// run the producer in the background
	go pizzeria(pizzaJob, pizzaTypes)

	// create and run consumer
	for i := range pizzaJob.data {
		if i.pizzaNumber <= NumberOfPizzas {
			if i.success {
				color.Green(i.message)
				color.Green("Order #%d is out for delivery!", i.pizzaNumber)
			} else {
				color.Red(i.message)
				color.Red("The customer is really mad!")
			}
		} else {
			color.Cyan("Done making pizzas...")
			err := pizzaJob.Close()
			if err != nil {
				color.Red("*** Error closing channel!", err)
			}
		}
	}

	// Wait for the pizzeria goroutine to finish
	pizzaJob.wg.Wait()

	// print out the ending message
	color.Cyan("-----------------")
	color.Cyan("Done for the day.")

	color.Cyan("We made %d pizzas, but failed to make %d, with %d attempts in total.", pizzasMade, pizzasFailed, total)

	switch {
	case pizzasFailed > 9:
		color.Red("It was an awful day...")
	case pizzasFailed >= 6:
		color.Red("It was not a very good day...")
	case pizzasFailed >= 4:
		color.Yellow("It was an okay day....")
	case pizzasFailed >= 2:
		color.Yellow("It was a pretty good day!")
	default:
		color.Green("It was a great day!")
	}
}
