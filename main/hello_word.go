package main
import (
	"fmt"
)

type HelloService struct {
	hello string
	hi string
}

func foo(s HelloService){
	s.hi = "new_hi"
	fmt.Println(s.hi)
}

func main() {
	s1 := new(HelloService)
	s1.hello = "hello"
	s1.hi = "hi"
	fmt.Println(s1)

	// print type of s1.hi
	fmt.Printf("%T\n", s1.hi)
	s2 := HelloService{"hello", "hi"}
	s2.hello = "hello1"
	s2.hi = "hi1"
	fmt.Println(s2)
	foo(s1)
	fmt.Println(s1)
	
}