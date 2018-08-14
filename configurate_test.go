package configurate

import (
	"testing"
	"strings"
	"os"
)

func TestDefaultsLoader(t *testing.T) {
	config := struct {
		Unconfigured string
		Used         string `default:"theDefault"`
		Overwritten  string `default:"theDefault"`
		Number       int    `default:"2"`
	}{
		Overwritten: "overwrittenValue",
	}

	err := NewDefaultsLoader().Load(&config)
	if err != nil {
		t.Fatalf("error %s", err.Error())
	}

	if config.Used != "theDefault" {
		t.Fatalf("did not load default but %q", config.Used)
	}

	if config.Overwritten != "overwrittenValue" {
		t.Fatalf("value was overwritten by default: %q", config.Overwritten)
	}

	if config.Number != 2 {
		t.Fatalf("Number != %d but %d", 2, config.Number)
	}
}

func TestJSONLoader(t *testing.T) {
	config := struct {
		FooBar string `json:"foo_bar"`
		Number int
	}{}
	json := `
{
  "foo_bar": "some_value",
  "number": 2
}
`
	err := NewJSONLoader(strings.NewReader(json)).Load(&config)
	if err != nil {
		t.Fatalf("error %s", err.Error())
	}

	if config.FooBar != "some_value" {
		t.Fatalf("Foobar != %q but %q", "some_value", config.FooBar)
	}

	if config.Number != 2 {
		t.Fatalf("Number != %d but %d", 2, config.Number)
	}
}

func TestEnvLoader(t *testing.T) {
	config := struct {
		Unconfigured string
		Used         string `env:"TEST_USED"`
		Number       int    `env:"TEST_INT"`
	}{}

	usedEnvBefore := os.Getenv("TEST_USED")
	numberEnvBefore := os.Getenv("TEST_INT")
	err := os.Setenv("TEST_USED", "env_value")
	if err != nil {
		t.Fatalf("error preparing env %s", err.Error())
	}
	err = os.Setenv("TEST_INT", "2")
	if err != nil {
		t.Fatalf("error preparing env %s", err.Error())
	}

	err = NewEnvLoader().Load(&config)
	if err != nil {
		t.Fatalf("error %s", err.Error())
	}

	if config.Used != "env_value" {
		t.Fatalf("did not load env but %q", config.Used)
	}

	if config.Number != 2 {
		t.Fatalf("Number != %d but %d", 2, config.Number)
	}

	err = os.Setenv("TEST_USED", usedEnvBefore)
	if err != nil {
		t.Fatalf("error resetting env %s", err.Error())
	}
	err = os.Setenv("TEST_INT", numberEnvBefore)
	if err != nil {
		t.Fatalf("error resetting env %s", err.Error())
	}
}
