package main

func Defer() {
	defer func() {
		println("第一个 defer")
	}()
	defer func() {
		println("第二个 defer")
	}()

	//println("第二个 defer")
	//println("第一个 defer")
}

func DeferLoop(max int) {
	for i := 0; i < max; i++ {
		defer func() {
			println("hello")
		}()
	}
}

func DeferClosure() {
	j := 0
	defer func() {
		println(j)
	}()
	j = 1
	//println(j)
}

func DeferClosureV1() {
	j := 0
	defer func(j int) {
		println(j)
	}(j)
	j = 1
}

func DeferReturn() int {
	a := 0
	defer func() {
		a = 1
	}()
	return a
}

func DeferReturnV1() (a int) {
	a = 0
	defer func() {
		a = 1
	}()
	return a
}

func DeferReturnV2() *MyStruct {
	res := &MyStruct{
		name: "Tom",
	}
	defer func() {
		res.name = "Jerry"
	}()
	return res
}

type MyStruct struct {
	name string
}
