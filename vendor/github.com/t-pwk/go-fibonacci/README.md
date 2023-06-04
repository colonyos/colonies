# High-performance Fibonacci numbers implementation in Go

[![Build Status](https://travis-ci.com/T-PWK/go-fibonacci.svg?branch=master)](https://travis-ci.com/T-PWK/go-fibonacci)
[![GitHub issues](https://img.shields.io/github/issues/T-PWK/go-fibonacci.svg)](https://github.com/T-PWK/go-fibonacci/issues)
[![Go Report Card](https://goreportcard.com/badge/github.com/T-PWK/go-fibonacci)](https://goreportcard.com/report/github.com/T-PWK/go-fibonacci)
[![Coverage Status](https://coveralls.io/repos/github/T-PWK/go-fibonacci/badge.svg?branch=master)](https://coveralls.io/github/T-PWK/go-fibonacci?branch=master)
[![GoDoc](https://godoc.org/github.com/T-PWK/go-fibonacci?status.svg)](https://godoc.org/github.com/T-PWK/go-fibonacci)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://blog.abelotech.com/mit-license/)

In mathematics, the Fibonacci numbers are the numbers in the following integer sequence, called the Fibonacci sequence, and characterized by the fact that every number after the first two is the sum of the two preceding ones:

```
1, 1, 2, 3, 5, 8, 13, 21, 34, 55, 89, 144, ...
```

Often, especially in modern usage, the sequence is extended by one more initial term:

```
0, 1, 1, 2, 3, 5, 8, 13, 21, 34, 55, 89, 144, ...
```

This implementation has two methods: `Fibonacci` and `FibonacciBig`. 

The `Fibonacci` function is more efficient, however, it returns correct numbers between 0 and 93 (inclusive). The `FibonacciBig` function, on the other hand, is less efficient but returns practically any Fibonacci number.

Example:

```go
package main

import (
  "fmt"
  "github.com/t-pwk/go-fibonacci"
)

func main() {

  fmt.Println("20: ", fib.Fibonacci(20))
  fmt.Println("200: ", fib.FibonacciBig(200))
}
```

And the output is

```
20:  6765
200:  280571172992510140037611932413038677189525
```
