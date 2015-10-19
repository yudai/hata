# 🏁 Hata - CLI Flags with Structures

Maybe something like below will be implemented.

```go
package main

import (
	"github.com/yudai/hata"
	"os"
)

type Options struct {
	ConfigFile string `short:"c", long:"config", desc:"Config file path"`
	User       string `short:"u", desc:"User name"`
	Timeout    uint   `short:"t", desc:"Timeout (seconds)"`
}

func main() {
	defaults := Options{
		ConfigFile: "/etc/foo/bar",
		User:       "admin",
		Timeout:    30,
	}

	c := hata.New(defaults)
	c.Parse(os.Args)
}

```
