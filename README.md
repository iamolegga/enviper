# enviper

[![PkgGoDev](https://pkg.go.dev/badge/github.com/iamolegga/enviper)](https://pkg.go.dev/github.com/iamolegga/enviper)
[![Build Status](https://circleci.com/gh/iamolegga/enviper.svg?style=svg)](https://circleci.com/gh/iamolegga/enviper)
[![Test Coverage](https://api.codeclimate.com/v1/badges/85fb13ce6638226a3732/test_coverage)](https://codeclimate.com/github/iamolegga/enviper/test_coverage)
[![Maintainability](https://api.codeclimate.com/v1/badges/85fb13ce6638226a3732/maintainability)](https://codeclimate.com/github/iamolegga/enviper/maintainability)
[![Go Report Card](https://goreportcard.com/badge/github.com/iamolegga/enviper)](https://goreportcard.com/report/github.com/iamolegga/enviper)

Package enviper is a helper/wrapper for [viper](http://github.com/spf13/viper) with the same API.
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
	"github.com/iamolegga/enviper"
	"github.com/spf13/viper"
)

type barry struct {
    Bar int `mapstructure:"bar"`
}
type bazzy struct {
    Baz bool
}
type quxxy struct {
	Qux string
}
type endpoint struct {
    Address string `mapstructure:"address"`
    Port    int    `mapstructure:"port"`
}
type config struct {
    Foo      string
    Barry    barry
    Barries  map[string]barry
    Endpoints []endpoint
    Bazzy    bazzy `mapstructure:",squash"`
    Quxxy    *quxxy
}

// For example this kind of structure can be unmarshaled with next yaml:
//  Foo: foo
//  Barry:
//    bar: 42
//  Baz: true
//  Barries: 
//    key1:
//      Bar: 255
//    key2:
//      Bar: 256
//  Endpoints:
//    - address: file-addr-0
//      port: 9000
//    - address: file-addr-1
//      port: 9001
//  Quxxy:
//    Qux: "lorem"
//
// And then it could be overriden by next env variables:
//  FOO=foo
//  BARRY_BAR=42
//  BAZ=true
//  BARRIES_KEY1_BAR=42
//  QUXXY_QUX=ipsum
//
// Slice elements are referenced by their index (Endpoints[0] and [1] above),
// each field of an element can be overridden separately:
//  ENDPOINTS_0_ADDRESS=env-addr-0
//  ENDPOINTS_0_PORT=9000
//  ENDPOINTS_1_PORT=9101
//
// Or with prefix:
//  MYAPP_FOO=foo
//  MYAPP_BARRY_BAR=42
//  MYAPP_BAZ=true
//  MYAPP_BARRIES_KEY1_BAR=42
//  MYAPP_ENDPOINTS_0_ADDRESS=env-addr-0
//  MYAPP_ENDPOINTS_1_PORT=9101
//  MYAPP_QUXXY_QUX=ipsum

func main() {    
    var c config

    e := enviper.New(viper.New())
    e.SetEnvPrefix("MYAPP")
    e.AddConfigPath("/my/config/path")
    e.SetConfigName("config")

    e.Unmarshal(&c)
}
```

## Slices & maps around env vars

Like viper (and like the maps above), a slice of objects can only be *overridden* from env vars: the slice skeleton has to be defined in the config file, and env vars can override fields of the existing elements by index. Env vars cannot create slice elements from scratch.

## Custom Tag Names

In case you want to use custom tag name (something different from `mapstructure`), you have to set it explicitly via `WithTagName` function.
The wrapper must know custom tag name in order to register all the env vars for viper so you can't just use `DecoderConfigOption`.

## Credits

Thanks to
[krak3n](https://github.com/krak3n) ([issuecomment-399884438](https://github.com/spf13/viper/issues/188#issuecomment-399884438))
and
[celian-garcia](https://github.com/celian-garcia) ([issuecomment-626122696](https://github.com/spf13/viper/issues/761#issuecomment-626122696))
for inspiring.
