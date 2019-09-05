# go-singleinstance

Cross plateform library to have only one instance of a software (based on python's [tendo](https://github.com/pycontribs/tendo/blob/master/tendo/singleton.py)).

## Usage

```go
package main

import (
	"fmt"
	"time"

	"github.com/allan-simon/go-singleinstance"
)

func main() {
	_, err := singleinstance.CreateLockFile("plop.lock")
	if err != nil {
		fmt.Println("An instance already exists")
		return
	}

	fmt.Println("Sleeping...")
	time.Sleep(30 * time.Second)
	fmt.Println("Done")
}
```

If you try to launch it twice, the second instance will fail.

## Thanks

For the python library trendo, from which I've shamelessly adapted the code.

## Contribution

Don't be afraid if it says "last commit 2 years ago", this library is made to be small
and simple so it's unlikely it changes after some times, however I'm pretty reactive
on github overall, so feel free to use issues to ask question, propose patch etc. :)

## License

MIT
