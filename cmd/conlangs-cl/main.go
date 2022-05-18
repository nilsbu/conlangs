package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/nilsbu/conlangs/pkg/creation"
	"github.com/nilsbu/conlangs/pkg/rand"
)

func main() {
	if def, err := os.ReadFile("test.def"); err != nil {
		fmt.Println(err)
	} else if creator, err := creation.NewCreator(def); err != nil {
		fmt.Println(err)
	} else {
		n, _ := strconv.Atoi(os.Args[1])
		rnd := rand.Natural(time.Now().UnixNano())
		for i := 0; i < n; i++ {
			fmt.Print(creator.Choose(rnd), " ")
		}
		fmt.Println()
	}
}
