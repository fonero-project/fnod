package fnoutil_test

import (
	"fmt"
	"math"

	"github.com/fonero-project/fnod/fnoutil"
)

func ExampleAmount() {

	a := fnoutil.Amount(0)
	fmt.Println("Zero Atom:", a)

	a = fnoutil.Amount(1e8)
	fmt.Println("100,000,000 Atoms:", a)

	a = fnoutil.Amount(1e5)
	fmt.Println("100,000 Atoms:", a)
	// Output:
	// Zero Atom: 0 FNO
	// 100,000,000 Atoms: 1 FNO
	// 100,000 Atoms: 0.001 FNO
}

func ExampleNewAmount() {
	amountOne, err := fnoutil.NewAmount(1)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(amountOne) //Output 1

	amountFraction, err := fnoutil.NewAmount(0.01234567)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(amountFraction) //Output 2

	amountZero, err := fnoutil.NewAmount(0)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(amountZero) //Output 3

	amountNaN, err := fnoutil.NewAmount(math.NaN())
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(amountNaN) //Output 4

	// Output: 1 FNO
	// 0.01234567 FNO
	// 0 FNO
	// invalid coin amount
}

func ExampleAmount_unitConversions() {
	amount := fnoutil.Amount(44433322211100)

	fmt.Println("Atom to kCoin:", amount.Format(fnoutil.AmountKiloCoin))
	fmt.Println("Atom to Coin:", amount)
	fmt.Println("Atom to MilliCoin:", amount.Format(fnoutil.AmountMilliCoin))
	fmt.Println("Atom to MicroCoin:", amount.Format(fnoutil.AmountMicroCoin))
	fmt.Println("Atom to Atom:", amount.Format(fnoutil.AmountAtom))

	// Output:
	// Atom to kCoin: 444.333222111 kFNO
	// Atom to Coin: 444333.222111 FNO
	// Atom to MilliCoin: 444333222.111 mFNO
	// Atom to MicroCoin: 444333222111 μFNO
	// Atom to Atom: 44433322211100 Atom
}
