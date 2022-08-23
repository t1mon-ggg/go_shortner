package main

import (
	"fmt"
	"os"
)

func mulfunc(i int) (int, error) {
	return i * 2, nil
}

func main() {
	// формулируем ожидания: анализатор должен находить ошибку,
	// описанную в комментарии want
	res, err := mulfunc(5)
	if err != nil {
		os.Exit(1) // want "os.Exit usage"
	}
	fmt.Println(res)
}
