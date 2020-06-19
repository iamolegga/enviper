package enviper_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/iamolegga/enviper"
)

func TestUnmarshal(t *testing.T) {
	suite.Run(t, new(UnmarshalSuite))
}

type UnmarshalSuite struct {
	suite.Suite
	v   *viper.Viper
	env map[string]string
}

func (s *UnmarshalSuite) SetupSuite() {}
func (s *UnmarshalSuite) SetupTest() {
	s.v = viper.New()
	s.setupFileConfig()
}
func (s *UnmarshalSuite) TearDownTest()  {}
func (s *UnmarshalSuite) TearDownSuite() {}

func (s *UnmarshalSuite) TestNotThrowsErrorWhenNoConfig() {
	var c Config
	e := enviper.New(viper.New())
	s.Nil(e.Unmarshal(&c))
	s.NotNil(c)
}

func (s *UnmarshalSuite) TestThrowsErrorWhenBrokenConfig() {
	var c Config
	v := viper.New()

	cwd, _ := os.Getwd()
	p := path.Join(cwd, "test_config.yaml")
	ioutil.WriteFile(p, []byte("qwerty"), 0777)
	defer os.Remove(p)

	e := enviper.New(v)
	e.AddConfigPath(cwd)
	e.SetConfigName("test_config")
	s.NotNil(e.Unmarshal(&c))
}

func (s *UnmarshalSuite) TestWithoutEnvs() {
	var c Config
	e := enviper.New(s.v)
	if err := e.Unmarshal(&c); err != nil {
		s.T().Error(err)
	}

	s.Equal("foo", c.Foo)
	s.Equal(1, c.Bar.BAZ)
	s.Equal(true, c.QUX.Quuux)
}

func (s *UnmarshalSuite) TestWithEnvs() {
	s.setupEnvConfig()

	var c Config
	e := enviper.New(s.v)
	e.SetEnvPrefix("PREF")

	if err := e.Unmarshal(&c); err != nil {
		s.T().Error(err)
	}

	s.Equal("fooooo", c.Foo)
	s.Equal(2, c.Bar.BAZ)
	s.Equal(false, c.QUX.Quuux)

	s.tearDownEnvConfig()
}

func (s *UnmarshalSuite) setupFileConfig() {
	cwd, _ := os.Getwd()
	s.v.AddConfigPath(cwd)
	s.v.SetConfigName("fixture")
}

func (s *UnmarshalSuite) setupEnvConfig() {
	s.loadEnvVars()

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
	Foo string
	Bar struct {
		BAZ int `mapstructure:"baz"`
	} `mapstructure:"bar"`
	QUX `mapstructure:",squash"`
}

type QUX struct {
	Quuux bool
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
	type config struct {
		Foo   string
		Barry barry
		Bazzy bazzy `mapstructure:",squash"`
	}

	// write config file

	dir := os.TempDir()
	defer os.RemoveAll(dir)
	p := path.Join(dir, "config.yaml")
	ioutil.WriteFile(p, []byte(`
Foo: foo
Barry:
  bar: 1
`), 0777)

	// write env vars that could override values from config file

	os.Setenv("MYAPP_BARRY_BAR", "2") // override value from file
	os.Setenv("MYAPP_BAZ", "false")
	defer os.Unsetenv("MYAPP_BARRY_BAR")
	defer os.Unsetenv("MYAPP_BAZ")

	// setup viper and enviper

	var c config
	e := enviper.New(viper.New())
	e.SetEnvPrefix("MYAPP")
	e.AddConfigPath(dir)
	e.SetConfigName("config")
	if err := e.Unmarshal(&c); err != nil {
		fmt.Printf("%+v\n", err)
	}

	fmt.Println(c.Foo)       // file only
	fmt.Println(c.Barry.Bar) // file & env, take env
	fmt.Println(c.Bazzy.Baz) // env only
	// Output:
	// foo
	// 2
	// false
}
