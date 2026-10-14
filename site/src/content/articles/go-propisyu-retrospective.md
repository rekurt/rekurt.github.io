# Russian number wording in Go: test the grammar and the precision

Turning a number into Russian words is not just a matter of replacing digits. A useful output depends on gender, noun forms and the way fractional values are handled. Those decisions deserve small, explicit examples before a library becomes part of a document-generation pipeline.

[go-propisyu](https://github.com/rekurt/go-propisyu) provides functions for number wording and related forms. This walkthrough uses v0.4.1, released on April 12, 2026. It is a retrospective on checks first executed September 18, with the pinned example rerun October 15.

## Keep the example versioned

Start in a fresh directory with Go installed:

```sh
mkdir propisyu-example
cd propisyu-example
go mod init example.com/propisyu-example
go get github.com/rekurt/go-propisyu@v0.4.1
```

Save the following as `main.go`, then run `go run .`:

```go
package main

import (
    "fmt"
    p "github.com/rekurt/go-propisyu"
)

func main() {
    cases := []struct{ got, want string }{
        {p.Money(1234, 56, p.CurrencyRUB),
            "одна тысяча двести тридцать четыре рубля пятьдесят шесть копеек"},
        {p.Ordinal(1, p.GenderFeminine), "первая"},
        {p.Decline(11, "заказ", "заказа", "заказов"), "заказов"},
        {p.Decline(21, "заказ", "заказа", "заказов"), "заказ"},
    }
    for _, c := range cases {
        if c.got != c.want {
            panic(fmt.Sprintf("got %q; want %q", c.got, c.want))
        }
        fmt.Println(c.got)
    }
    got, err := p.DecimalToWordsPrecision("3.14159", 3)
    if err != nil || got != "три целых сто сорок одна тысячная" {
        panic(fmt.Sprintf("unexpected decimal: %q (%v)", got, err))
    }
    fmt.Println(got)
}
```

The example checks strings rather than merely printing output for visual inspection. If one of the expected forms changes, it stops with an error.

## Three different decisions in one short program

`Money` receives the major and minor units separately in this example. The output includes both rubles and kopecks, with the forms appropriate to these values. This demonstrates the formatting call; it does not define how an application should calculate or round monetary amounts.

`Ordinal` makes the grammatical gender explicit. Here the expected feminine first ordinal is `первая`. A caller needs to choose the form that matches the surrounding text.

`Decline` chooses among the three supplied forms. Eleven and twenty-one are useful contrasting inputs: this example expects `заказов` for 11 and `заказ` for 21. The library selects a form; the application supplies the nouns.

## Precision is not rounding

For the decimal input `3.14159` and precision 3, the checked output describes 3.141. The function truncates the remaining digits in this case. It does not produce the wording for the rounded value 3.142.

That distinction belongs in the calling code's design. If the application requires rounding, establish that numerical rule before passing the value to a wording function, and test the result independently. The string input makes this example's decimal value visible without introducing an intermediate binary floating-point literal.

## Turn the example into application tests

The useful next step is a small table of the exact phrases your application needs: its currencies, genders, noun forms and decimal policies. Keep expected text beside the input. Include the boundary values that matter to that application rather than relying on one successful example.

The pinned program above passed locally on macOS arm64. A September source-checkout test run also passed, but its coverage percentage is not a promise about every grammatical case or future version. This article does not validate legal document requirements or the numerical correctness of a larger accounting workflow.

[Source and API documentation](https://github.com/rekurt/go-propisyu) · [v0.4.1 release](https://github.com/rekurt/go-propisyu/releases/tag/v0.4.1)

Disclosure: this article was drafted with AI assistance. The pinned Go example and its assertions were executed locally.
