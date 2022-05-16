package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/nilsbu/conlangs/pkg/genesis"
)

func main() {
	if def, err := os.ReadFile("test.def"); err != nil {
		fmt.Println(err)
	} else if creator, err := genesis.NewCreator(def); err != nil {
		fmt.Println(err)
	} else {
		n, _ := strconv.Atoi(os.Args[1])
		cnt := creator.N()
		rand.Seed(time.Now().UnixNano())
		for i := 0; i < n; i++ {
			r := rand.Int() % cnt
			fmt.Println(creator.Get(r))
		}
	}
}
