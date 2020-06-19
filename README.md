# enviper

[![GoDoc](https://godoc.org/github.com/iamolegga/enviper?status.svg)](https://godoc.org/github.com/iamolegga/enviper)
[![Maintainability](https://api.codeclimate.com/v1/badges/85fb13ce6638226a3732/maintainability)](https://codeclimate.com/github/iamolegga/enviper/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/85fb13ce6638226a3732/test_coverage)](https://codeclimate.com/github/iamolegga/enviper/test_coverage)
[![Go Report Card](https://goreportcard.com/badge/github.com/iamolegga/enviper)](https://goreportcard.com/report/github.com/iamolegga/enviper)

Enviper is a helper/wrapper package for [viper](http://github.com/spf13/viper) with the same API.
It makes it possible to unmarshal config to struct considering environment variables.

## Problem

[Viper](https://github.com/spf13/viper) package doesn't consider environment variables while unmarshaling.
Please, see: [188](https://github.com/spf13/viper/issues/188) and [761](https://github.com/spf13/viper/issues/761)

## Solution

Just wrap viper instance and use the same `Unmarshal` method as you did before:

```go
e := enviper.New(viper.New())
e.Unmarshal(&config)
```

## Example

```go
package main

import (
	"github.com/spf13/viper"
	"github.com/iamolegga/enviper"
)

type barry struct {
    Bar int `mapstructure:"bar"`
}
type bazzy struct {
    Baz bool
}
type config struct {
    Foo string
    Barry barry
    Bazzy bazzy `mapstructure:",squash"`
}

func main() {    
    var c config

    e := enviper.New(viper.New())
    e.SetEnvPrefix("MYAPP")
    e.AddConfigPath("/my/config/path")
    e.SetConfigName("config")

    e.Unmarshal(&c)
}
```

## Credits

Thanks to
[krak3n](https://github.com/krak3n) ([issuecomment-399884438](https://github.com/spf13/viper/issues/188#issuecomment-399884438))
and
[celian-garcia](https://github.com/celian-garcia) ([issuecomment-626122696](https://github.com/spf13/viper/issues/761#issuecomment-626122696))
for inspiring.
