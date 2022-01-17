package enviper

import (
	"github.com/mitchellh/mapstructure"
	"reflect"
	"strings"

	"github.com/spf13/viper"
)

// Enviper is a wrapper struct for viper,
// that makes it possible to unmarshal config to struct
// considering environment variables
type Enviper struct {
	*viper.Viper
	tagName string
}

// New returns an initialized Enviper instance
func New(v *viper.Viper) *Enviper {
	return &Enviper{
		Viper: v,
	}
}

const defaultTagName = "mapstructure"

// WithTagName sets custom tag name to be used instead of default `mapstructure`
func (e *Enviper) WithTagName(customTagName string) *Enviper {
	e.tagName = customTagName
	return e
}

// TagName returns currently used tag name (`mapstructure` by default)
func (e *Enviper) TagName() string {
	if e.tagName == "" {
		return defaultTagName
	}
	return e.tagName
}

// Unmarshal unmarshals the config into a Struct just like viper does.
// The difference between enviper and viper is in automatic overriding data from file by data from env variables
func (e *Enviper) Unmarshal(rawVal interface{}, opts ...viper.DecoderConfigOption) error {
	if e.TagName() != defaultTagName {
		opts = append(opts, func(c *mapstructure.DecoderConfig) {
			c.TagName = e.TagName()
		})
	}

	if err := e.Viper.ReadInConfig(); err != nil {
		switch err.(type) {
		case viper.ConfigFileNotFoundError:
			// 	do nothing
		default:
			return err
		}
	}
	// We need to unmarshal before the env binding to make sure that keys of maps are bound just like the struct fields
	// We silence errors here because we'll unmarshal a second time
	_ = e.Viper.Unmarshal(rawVal, opts...)
	e.readEnvs(rawVal)
	return e.Viper.Unmarshal(rawVal, opts...)
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
		if fv.Kind() == reflect.Ptr {
			if fv.IsZero() {
				fv = reflect.New(fv.Type().Elem()).Elem()
			} else {
				fv = fv.Elem()
			}
		}
		t := ifv.Type().Field(i)
		tv, ok := t.Tag.Lookup(e.TagName())
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
		case reflect.Map:
			iter := fv.MapRange()
			for iter.Next() {
				if key, ok := iter.Key().Interface().(string); ok {
					e.bindEnvs(iter.Value().Interface(), append(prev, tv, key)...)
				}
			}
		default:
			env := strings.Join(append(prev, tv), ".")
			// Viper.BindEnv will never return error
			// because env is always non empty string
			_ = e.Viper.BindEnv(env)
		}
	}
}
