package main

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

func sampleMain() {
	var name int = 10
	const myConst string = "abc"
	fmt.Println(myConst)
	fmt.Println(name)

	stringsAndRunes()
}

func fnName(printValue string) {
	fmt.Println("running function")
	fmt.Println(printValue)
}

func division(numerator int, denom int) (int, int, error) {
	var err error = nil
	if denom == 0 {
		err = errors.New("Error denom 0")
	}
	return numerator / denom, numerator % denom, err
}

func arrayGo() {
	var intArr [3]int = [3]int{1, 2, 4}
	t := [...]int{1, 24}

	var arrDynamic []int32 = make([]int32, 3 /*size*/, 8 /*capacity*/)

	arrDynamic2 := []int32{4, 5, 6}

	arrDynamic = append(arrDynamic, 10)
	arrDynamic = append(arrDynamic, arrDynamic2...)

	len := len(intArr)
	fmt.Println(len)

	fmt.Println(intArr[2])
	fmt.Println(intArr[1:3])

	t[1] = 10

	intArr[0] = 5

	fmt.Println(&intArr[2])
}

func mapInGo() {
	var myMap map[string]int = make(map[string]int)

	var map2 = map[string]int{"Abc": 3, "dada": 93}

	fmt.Println(len(myMap))
	var value, exists = map2["Abc"]

	fmt.Println(value)
	fmt.Println(exists)

}

func forLoop(myMap map[string]int, myArray []int32) {

	for key, value := range myMap {
		fmt.Printf("Key = %v, Value = %v\n", key, value)
	}

	for i, j := range myArray {
		fmt.Printf("Index = %v, Value=%v", i, j)
	}

	for i := 0; i <= 10; i++ {
		fmt.Println(i)
	}

	k := 0
	for {
		k++
		fmt.Print(k)
		if k > 5 {
			break
		}
	}
}

func stringsAndRunes() {
	var str string = "résumé"

	// Go uses Utf 8 encoding for strings

	fmt.Println(len(str)) // returns 8 because of é(2 each)

	for index, char := range str {
		fmt.Printf("Index = %v, Char = %v", index, char)
	}

	// rune -> 32 bits so no issues in this
	var char rune = 'a'
	fmt.Println(char)

	// simple way for a string - []rune
	newString := []rune{'a', 'é', 'c'}
	fmt.Println(newString[0])

	var builder strings.Builder
	for i := range str {
		builder.WriteByte(str[i])
	}

	fmt.Println(builder.String())

}

/*
the make keyword is used specifically to allocate and initialize memory for slices, maps, and channels.
make([]T, length, capacity)
make(map[keyType]valueType)
make(chan T, bufferSize)
*/

// defer panic recover [exception handling]

func concurrency() {
	tasks := []string{"task1", "task2", "task3"}
	for _, task := range tasks {
		performTask(task)
	}
}

func performTask(task string) {
	fmt.Println(task)
	time.Sleep(time.Second)
}
