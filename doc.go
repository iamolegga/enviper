// Package enviper is a helper/wrapper for http://github.com/spf13/viper with the same API.
// It makes it possible to unmarshal config to struct
// considering environment variables.
//
// Problem
//
// Viper package (https://github.com/spf13/viper) doesn't consider environment variables while unmarshaling.
// Please, see:
// https://github.com/spf13/viper/issues/188
// and
// https://github.com/spf13/viper/issues/761
//
// Solution
//
// Just wrap viper instance and use the same `Unmarshal` method as you did before:
//
// 	e := enviper.New(viper.New())
// 	e.Unmarshal(&config)
//
// Credits
//
// Thanks to https://github.com/krak3n (https://github.com/spf13/viper/issues/188#issuecomment-399884438)
// and
// https://github.com/celian-garcia (https://github.com/spf13/viper/issues/761#issuecomment-626122696)
// for inspiring.
package enviper
