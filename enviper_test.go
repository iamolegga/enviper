package enviper_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/iamolegga/enviper"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestUnmarshal(t *testing.T) {
	suite.Run(t, new(UnmarshalSuite))
}

type UnmarshalSuite struct {
	suite.Suite
	v   *viper.Viper
	env map[string]string
}

func (s *UnmarshalSuite) SetupSuite() {
	s.loadEnvVars()
}
func (s *UnmarshalSuite) SetupTest() {
	s.v = viper.New()
}
func (s *UnmarshalSuite) TearDownTest()  {}
func (s *UnmarshalSuite) TearDownSuite() {}

func (s *UnmarshalSuite) TestThrowsErrorWhenBrokenConfig() {
	cwd, _ := os.Getwd()
	p := path.Join(cwd, "test_config.yaml")
	ioutil.WriteFile(p, []byte("qwerty"), 0777)
	defer os.Remove(p)

	var c Config
	e := enviper.New(s.v)
	e.AddConfigPath(cwd)
	e.SetConfigName("test_config")
	s.NotNil(e.Unmarshal(&c))
}

func (s *UnmarshalSuite) TestOnlyEnvsByCustomTag() {
	s.setupEnvConfig()
	defer s.tearDownEnvConfig()

	var c Config
	e := enviper.New(s.v).WithTagName("custom_tag")
	e.SetEnvPrefix("PREF")
	s.Nil(e.Unmarshal(&c))

	s.Equal("imTheValueByCustomTag", c.TagTest)
}

func (s *UnmarshalSuite) TestOnlyEnvsWithTagValue() {
	s.setupEnvConfig()
	defer s.tearDownEnvConfig()

	var c Config
	e := enviper.New(s.v)
	e.SetEnvPrefix("PREF")
	s.Nil(e.Unmarshal(&c))

	s.Equal("imTheValueWithOmitempty", c.TagValueWithOmitempty)
	s.Equal("imTheValueWithCustomName", c.TagValueWithNameOmitempty)
	s.Equal("", c.TagValueWithDash)
}

func (s *UnmarshalSuite) TestOnlyEnvs() {
	s.setupEnvConfig()
	defer s.tearDownEnvConfig()

	var c Config
	e := enviper.New(s.v)
	e.SetEnvPrefix("PREF")
	s.Nil(e.Unmarshal(&c))

	s.Equal("fooooo", c.Foo)
	s.Equal(2, c.Bar.BAZ)
	s.Equal(false, c.QUX.Quuux)
	s.Equal("testptr3", c.QUX.QuuuxPtrUnset.Value)

	// TODO: known bug, that maps can not be set when there is no config file. Uncomment and fix
	//s.Equal(true, c.QuxMap["key1"].Quuux)
}

func (s *UnmarshalSuite) TestPrimitiveMap() {
	s.setupFileConfig()

	var c Config
	e := enviper.New(s.v)
	s.Nil(e.Unmarshal(&c))

	s.Equal("val1", c.PrimitiveMap["key1"])
	s.Equal("val2", c.PrimitiveMap["key2"])
	s.Equal("", c.PrimitiveMap[""])
}

func (s *UnmarshalSuite) TestOnlyConfig() {
	s.setupFileConfig()

	var c Config
	e := enviper.New(s.v)
	s.Nil(e.Unmarshal(&c))

	s.Equal("foo", c.Foo)
	s.Equal(1, c.Bar.BAZ)
	s.Equal(true, c.QUX.Quuux)
	s.Equal("testptr1", c.FooPtr.Value)
	s.Equal("testptr2", c.QUX.QuuuxPtr.Value)
	s.Len(c.Storage.S3, 2)
	s.Equal("file-addr-0", c.Storage.S3[0].Address)
	s.Equal(9000, c.Storage.S3[0].Port)
	s.Equal("file-addr-1", c.Storage.S3[1].Address)
	s.Equal(9001, c.Storage.S3[1].Port)
}

func (s *UnmarshalSuite) TestConfigWithEnvs() {
	s.setupFileConfig()
	s.setupEnvConfig()
	defer s.tearDownEnvConfig()

	var c Config
	e := enviper.New(s.v)
	e.SetEnvPrefix("PREF")
	s.Nil(e.Unmarshal(&c))

	s.Equal("fooooo", c.Foo)
	s.Equal(2, c.Bar.BAZ)
	s.Equal(false, c.QUX.Quuux)
	s.Equal(true, c.QuxMap["key1"].Quuux)
	s.Equal(false, c.QuxMap["key2"].Quuux)
	s.Equal("testptr1", c.FooPtr.Value)
	s.Equal("testptr2", c.QUX.QuuuxPtr.Value)
	s.Equal("testptr3", c.QuuuxPtrUnset.Value)
}

func (s *UnmarshalSuite) TestSliceOfObjectsOverriddenByEnvs() {
	s.setupFileConfig()
	s.setupEnvConfig()
	defer s.tearDownEnvConfig()

	var c Config
	e := enviper.New(s.v)
	e.SetEnvPrefix("PREF")
	s.Nil(e.Unmarshal(&c))

	// env vars reference slice elements by index: PREF_STORAGE_S3_<i>_<field>.
	// Index 0: address overridden by env, port stays from the config file.
	s.Equal("env-addr-0", c.Storage.S3[0].Address)
	s.Equal(9000, c.Storage.S3[0].Port)
	// Index 1: both address and port overridden by env.
	s.Equal("env-addr-1", c.Storage.S3[1].Address)
	s.Equal(9101, c.Storage.S3[1].Port)
}

func (s *UnmarshalSuite) TestSliceOfObjectsEnvOnlyStaysEmpty() {
	s.setupEnvConfig()
	defer s.tearDownEnvConfig()

	var c Config
	e := enviper.New(s.v)
	e.SetEnvPrefix("PREF")
	s.Nil(e.Unmarshal(&c))

	// Known limitation, same as for maps: env vars can override slice elements
	// defined by the config file, but they cannot create the slice from scratch,
	// because the slice skeleton only exists after the config file is unmarshaled.
	s.Empty(c.Storage.S3)
}

func (s *UnmarshalSuite) setupFileConfig() {
	cwd, _ := os.Getwd()
	s.v.AddConfigPath(cwd)
	s.v.SetConfigName("fixture")
}

func (s *UnmarshalSuite) setupEnvConfig() {
	for k, v := range s.env {
		if err := os.Setenv(k, v); err != nil {
			s.T().Error(err)
		}
	}
}

func (s *UnmarshalSuite) tearDownEnvConfig() {
	for k := range s.env {
		if err := os.Unsetenv(k); err != nil {
			s.T().Error(err)
		}
	}
}

func (s *UnmarshalSuite) loadEnvVars() {
	cwd, _ := os.Getwd()
	p := path.Join(cwd, "fixture_env")
	bytes, err := ioutil.ReadFile(p)
	if err != nil {
		s.T().Error(err)
	}
	content := string(bytes)
	raws := strings.Split(content, "\n")
	s.env = make(map[string]string, len(raws))
	for _, raw := range raws {
		if len(raw) == 0 {
			continue
		}
		pair := strings.Split(raw, "=")
		if len(pair) != 2 {
			s.T().Error(errors.New("invalid env fixtures"))
		}
		k := pair[0]
		v := pair[1]
		s.env[k] = v
	}
}

type Config struct {
	Foo    string
	FooPtr *PtrTest
	Bar    struct {
		BAZ int `mapstructure:"baz"`
	} `mapstructure:"bar"`
	QuxMap map[string]struct {
		Quuux bool
	}
	PrimitiveMap map[string]string

	Storage StorageConfig `mapstructure:"storage"`

	QUX                       `mapstructure:",squash"`
	TagTest                   string `custom_tag:"TAG_TEST"`
	TagValueWithOmitempty     string `mapstructure:",omitempty"`
	TagValueWithNameOmitempty string `mapstructure:"tag_custom_name,omitempty"`
	TagValueWithDash          string `mapstructure:"-"`
}

type StorageConfig struct {
	S3 []S3Config `mapstructure:"s3"`
}

type S3Config struct {
	Address string `mapstructure:"address"`
	Port    int    `mapstructure:"port"`
}

type QUX struct {
	Quuux         bool
	QuuuxPtr      *PtrTest `mapstructure:"quuux_ptr"`
	QuuuxPtrUnset *PtrTest `mapstructure:"quuux_ptr_unset"`
}

type PtrTest struct {
	Value string
}

func TestNew(t *testing.T) {
	v := viper.New()
	assert.Exactly(t, &enviper.Enviper{Viper: v}, enviper.New(v))
}

func ExampleEnviper_Unmarshal() {
	// describe config structure

	type barry struct {
		Bar int `mapstructure:"bar"`
	}
	type bazzy struct {
		Baz bool
	}
	type endpoint struct {
		Address string `mapstructure:"address"`
		Port    int    `mapstructure:"port"`
	}
	type config struct {
		Foo       string
		Barry     barry
		Bazzy     bazzy      `mapstructure:",squash"`
		Endpoints []endpoint `mapstructure:"endpoints"`
	}

	// write config file

	dir := os.TempDir()
	defer os.RemoveAll(dir)
	p := path.Join(dir, "config.yaml")
	ioutil.WriteFile(p, []byte(`
Foo: foo
Barry:
  bar: 1
Endpoints:
  - address: file-addr-0
    port: 9000
  - address: file-addr-1
    port: 9001
`), 0777)

	// write env vars that could override values from config file.
	// slice elements are referenced by their index:
	// MYAPP_ENDPOINTS_<index>_<field>

	os.Setenv("MYAPP_BARRY_BAR", "2") // override value from file
	os.Setenv("MYAPP_BAZ", "false")
	os.Setenv("MYAPP_ENDPOINTS_0_ADDRESS", "env-addr-0") // override address of the 1st element
	os.Setenv("MYAPP_ENDPOINTS_1_PORT", "9101")          // override port of the 2nd element
	defer os.Unsetenv("MYAPP_BARRY_BAR")
	defer os.Unsetenv("MYAPP_BAZ")
	defer os.Unsetenv("MYAPP_ENDPOINTS_0_ADDRESS")
	defer os.Unsetenv("MYAPP_ENDPOINTS_1_PORT")

	// setup viper and enviper

	var c config
	e := enviper.New(viper.New())
	e.SetEnvPrefix("MYAPP")
	e.AddConfigPath(dir)
	e.SetConfigName("config")
	if err := e.Unmarshal(&c); err != nil {
		fmt.Printf("%+v\n", err)
	}

	fmt.Println(c.Foo)                  // file only
	fmt.Println(c.Barry.Bar)            // file & env, take env
	fmt.Println(c.Bazzy.Baz)            // env only
	fmt.Println(c.Endpoints[0].Address) // file & env, take env
	fmt.Println(c.Endpoints[0].Port)    // file only
	fmt.Println(c.Endpoints[1].Address) // file only
	fmt.Println(c.Endpoints[1].Port)    // file & env, take env
	// Output:
	// foo
	// 2
	// false
	// env-addr-0
	// 9000
	// file-addr-1
	// 9101
}
