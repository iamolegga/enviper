package enviper

import (
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-viper/mapstructure/v2"

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

	switch ifv.Kind() {
	case reflect.Struct:
		for i := 0; i < ifv.NumField(); i++ {
			fv := ifv.Field(i)
			if fv.Kind() == reflect.Ptr {
				if fv.IsZero() {
					fv = reflect.New(fv.Type().Elem()).Elem()
				} else {
					fv = fv.Elem()
				}
			}

			seg, squash, ok := structFieldSeg(ifv.Type(), i, e.TagName())
			if !ok {
				continue
			}

			// If "squash" is specified in the tag, we squash the field down.
			if squash {
				e.bindEnvs(fv.Interface(), prev...)
				continue
			}

			e.bindEnvs(fv.Interface(), append(prev, seg)...)
		}
	case reflect.Map:
		iter := ifv.MapRange()
		for iter.Next() {
			if key, ok := iter.Key().Interface().(string); ok {
				e.bindEnvs(iter.Value().Interface(), append(prev, key)...)
			}
		}
	case reflect.Slice, reflect.Array:
		e.bindEnvSlice(ifv, prev)
	default:
		env := strings.Join(prev, ".")
		// Viper.BindEnv will never return error
		// because env is always non empty string
		_ = e.Viper.BindEnv(env)
	}
}

// structFieldSeg resolves the path segment for the i-th field of a struct type,
// following the same tag semantics as bindEnvs: a "-" value skips the field, the
// "squash" option flattens it, and an empty tag falls back to the field name.
// It returns ok=false when the field must be skipped entirely.
func structFieldSeg(st reflect.Type, i int, tagName string) (seg string, squash bool, ok bool) {
	tv, hasTag := st.Field(i).Tag.Lookup(tagName)
	if !hasTag {
		tv = st.Field(i).Name
	} else if index := strings.Index(tv, ","); index != -1 {
		if tv[:index] == "-" {
			return "", false, false
		}

		// If "squash" is specified in the tag, we squash the field down.
		if strings.Contains(tv[index+1:], "squash") {
			return "", true, true
		}

		tv = tv[:index]
	}

	if tv == "" {
		tv = st.Field(i).Name
	}

	return tv, false, true
}

// bindEnvSlice binds environment variables for the elements of a slice field.
// Slice elements are referenced by their index: the address of the first element
// of `storage.s3` maps to `PREF_STORAGE_S3_0_ADDRESS`. Unlike maps and scalars —
// which are registered through viper's BindEnv — slice elements are merged into a
// single override value set for the whole slice key via Viper.Set. Viper's
// settings builder (getSettings/deepSearch) cannot combine a config-file array
// (kept as a whole value at `storage.s3`) with flat env keys such as
// `storage.s3.0.address`, so registering the latter with BindEnv must be avoided.
func (e *Enviper) bindEnvSlice(sv reflect.Value, prev []string) {
	if sv.Len() == 0 {
		// Same limitation as maps: the slice skeleton has to come from the config
		// file (populated by the first unmarshal), env vars can only override
		// existing elements.
		return
	}

	e.Viper.Set(strings.Join(prev, "."), e.envMergedValue(sv, prev))
}

// envMergedValue returns the value for the subtree rooted at v, where env vars
// (addressed by the path prev) override the config-file values already present in
// v from the first unmarshal. It mirrors the traversal rules of bindEnvs.
func (e *Enviper) envMergedValue(v reflect.Value, prev []string) any {
	switch v.Kind() {
	case reflect.Interface, reflect.Ptr:
		if v.IsNil() {
			return nil
		}
		return e.envMergedValue(v.Elem(), prev)
	case reflect.Struct:
		m := make(map[string]any)
		for i := 0; i < v.NumField(); i++ {
			fv := v.Field(i)
			if fv.Kind() == reflect.Ptr {
				// Mirror bindEnvs: dereference named pointers, allocating a zero
				// value if it is nil so that env vars can still fill it.
				if fv.IsNil() {
					fv = reflect.New(fv.Type().Elem())
				}
				fv = fv.Elem()
			}

			seg, squash, ok := structFieldSeg(v.Type(), i, e.TagName())
			if !ok {
				continue
			}

			if squash {
				// Squash flattens the sub-struct into the current level.
				sub := e.envMergedValue(fv, prev)
				for k, subv := range sub.(map[string]any) {
					m[k] = subv
				}
				continue
			}

			m[seg] = e.envMergedValue(fv, append(prev, seg))
		}
		return m
	case reflect.Map:
		m := make(map[string]any)
		iter := v.MapRange()
		for iter.Next() {
			if key, ok := iter.Key().Interface().(string); ok {
				m[key] = e.envMergedValue(iter.Value(), append(prev, key))
			}
		}
		return m
	case reflect.Slice, reflect.Array:
		if v.Len() == 0 {
			return v.Interface()
		}
		out := make([]any, v.Len())
		for i := 0; i < v.Len(); i++ {
			out[i] = e.envMergedValue(v.Index(i), append(prev, strconv.Itoa(i)))
		}
		return out
	default:
		key := strings.Join(prev, ".")
		if val, ok := e.envValue(key); ok {
			return val
		}
		return v.Interface()
	}
}

// envValue returns the value of the environment variable that viper would use for
// the given key: the configured env prefix plus the key, upper-cased, with dots
// replaced by underscores (the env key replacer set up by readEnvs).
func (e *Enviper) envValue(key string) (string, bool) {
	var name string
	if prefix := e.Viper.GetEnvPrefix(); prefix != "" {
		name = prefix + "_" + key
	} else {
		name = key
	}
	name = strings.ToUpper(name)
	name = strings.NewReplacer(".", "_").Replace(name)

	val, ok := os.LookupEnv(name)
	return val, ok && val != ""
}
