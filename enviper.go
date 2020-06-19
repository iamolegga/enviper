// Enviper package is a helper/wrapper for http://github.com/spf13/viper with the same API.
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

import (
	"reflect"
	"strings"

	"github.com/spf13/viper"
)

// Enviper is a wrapper struct for viper,
// that makes it possible to unmarshal config to struct
// considering environment variables
type Enviper struct {
	*viper.Viper
}

// New returns an initialized Enviper instance
func New(v *viper.Viper) *Enviper {
	return &Enviper{v}
}

// Unmarshal unmarshals the config into a Struct just like viper does.
// The difference between enviper and viper is in automatic overriding data from file by data from env variables
func (e *Enviper) Unmarshal(rawVal interface{}) error {
	if err := e.Viper.ReadInConfig(); err != nil {
		return err
	}
	e.readEnvs(rawVal)
	return e.Viper.Unmarshal(rawVal)
}

func (e *Enviper) readEnvs(rawVal interface{}) {
	e.Viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	e.bindEnvs(rawVal)
}

func (e *Enviper) bindEnvs(in interface{}, prev ...string) {
	ifv := reflect.ValueOf(in)
	if ifv.Kind() == reflect.Ptr {
		ifv = ifv.Elem()
	}
	for i := 0; i < ifv.NumField(); i++ {
		fv := ifv.Field(i)
		t := ifv.Type().Field(i)
		tv, ok := t.Tag.Lookup("mapstructure")
		if ok {
			if tv == ",squash" {
				e.bindEnvs(fv.Interface(), prev...)
				continue
			}
		} else {
			tv = t.Name
		}
		switch fv.Kind() {
		case reflect.Struct:
			e.bindEnvs(fv.Interface(), append(prev, tv)...)
		default:
			env := strings.Join(append(prev, tv), ".")
			// Viper.BindEnv will never return error
			// because env is always non empty string
			_ = e.Viper.BindEnv(env)
		}
	}
}
