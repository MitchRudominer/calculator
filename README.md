# calculator
A scanner and recursive-descent parser written in the Go language.

Used to implement an integer calculator that handles addition, subtraction and multiplication.
Uses Go's `big.Int` class for arbitrarily big integer arithmetic.

Build and run:
```
go install
calculator
Enter an arithmetic expression: 1234567890 * 2345678901 * (3456789023 + 4567890123) - 5678901234
23238667146635409191100386706
Enter an arithmetic expression: (5 + 7) * ( (3) * (6 + 1) + 1)
264
Enter an arithmetic expression: (5 + 7) * ( (3) * 6 + 1) + 1)
Unexpected token at postion 28: )

```

Run the tests:
```
cd parser
go test
```
