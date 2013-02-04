package yail

import "os"

func ExampleHelloWorld() {
	runExample(`@println("Labas, pasauli!")`)
	// Output: Labas, pasauli!
}

func ExampleFor() {
	runExample(`for i = 0; i < 5; i = i + 2 { @println(i, i + 1) }`)
	// Output: 0 1
	// 2 3
	// 4 5
}

func ExampleFactorial() {
	runExample(`fun = (x) {
		fun = .fun
		if x == 0 {
			return 1
		} else {
			return x * fun(x - 1)
		}
	}
	@print(fun(5))`)
	// Output: 120
}

func ExamplePrimes() {
	runExample(`MAX = 100
	for i = 2; i < MAX; i = i + 1 {
		isPrime[i] = true
	}
	numPrimes = 0
	for i = 2; i < MAX; i = i + 1 {
		if isPrime[i] {
			primes[numPrimes] = i
			numPrimes = numPrimes + 1
			for j = i * 2; j < MAX; j = j + i {
				isPrime[j] = false
			}
		}
	}
	for i = 0; i < numPrimes; i = i + 1 {
		@println(primes[i])
	}`)
	// Output: 2
	// 3
	// 5
	// 7
	// 11
	// 13
	// 17
	// 19
	// 23
	// 29
	// 31
	// 37
	// 41
	// 43
	// 47
	// 53
	// 59
	// 61
	// 67
	// 71
	// 73
	// 79
	// 83
	// 89
	// 97
}

func runExample(source string) {
	Interpret(source, os.Stdin, os.Stdout)
}
