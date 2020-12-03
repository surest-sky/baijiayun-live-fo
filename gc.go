package main

import (
	"fmt"
)

func add(a, b int, ch chan int) {
	c := a + b
	fmt.Printf("%d + %d = %d\n", a, b, c)
	ch <- 1
}


//func main() {
//	start := time.Now()
//	chs := make([]chan int, 10)
//	for i := 0; i < 10; i++ {
//		chs[i] = make(chan int)
//		go add(1, i, chs[i])
//	}
//	for _, ch := range chs {
//		<- ch
//	}
//	end := time.Now()
//	consume := end.Sub(start).Seconds()
//	fmt.Println("程序执行耗时(s)：", consume)
//}

func p(i chan class)  {
	testClas := class{name:"surest", age:156}
	i <- testClas
}

type class struct {
	name string
	age int
}

func main()  {
	t := 10
	testChans := make([]chan class, t)
	for i:=0; i < 10; i++ {
		testChans[i] = make(chan class)
		go func(i int) {
			p(testChans[i])
		}(i)
	}

	for _, testChan := range testChans {
		fmt.Println(<- testChan)
	}
}
