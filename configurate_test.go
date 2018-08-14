package configurate

import (
	"testing"
	"strings"
	"os"
)

func TestLoadFile(t *testing.T) {
	config := struct {
		AppName           string  `json:"app_name"`
		NumberOfRetries   int     `json:"number_of_retries"`
		Version           string  `json:"version" default:"1.0"`
		SomethingOptional *string `json:"something_optional"`
	}{}

	err := LoadFile("config.json", &config)
	if err != nil {
		t.Fatalf("error %s", err.Error())
	}

	if config.AppName != "ConfigurateJson" {
		t.Fatalf("AppName not as specified but: %q", config.AppName)
	}

	if config.NumberOfRetries != 5 {
		t.Fatalf("NumberOfRetries not as specified but: %d", config.NumberOfRetries)
	}

	if config.Version != "1.0" {
		t.Fatalf("Version not as specified but: %d", config.Version)
	}

	if config.SomethingOptional != nil {
		t.Fatal("SomethingOptional is not nil")
	}
}

func TestLoadFileWithOptional(t *testing.T) {
	config := struct {
		AppName           string  `json:"app_name"`
		NumberOfRetries   int     `json:"number_of_retries"`
		Version           string  `json:"version" default:"1.0"`
		SomethingOptional *string `json:"something_optional"`
	}{}

	err := LoadFile("config2.json", &config)
	if err != nil {
		t.Fatalf("error %s", err.Error())
	}

	if config.SomethingOptional == nil {
		t.Fatal("SomethingOptional is nil")
	}

	if *config.SomethingOptional != "the value" {
		t.Fatalf("SomethingOptional is not as expected. is: %q", *config.SomethingOptional)
	}
}

func TestLoadAll(t *testing.T) {
	reverter, err := changeEnv(map[string]string{
		"NUMBER_OF_RETRIES": "12",
	})
	if err != nil {
		t.Fatalf("error preparing env %s", err.Error())
	}
	defer reverter()

	json := strings.NewReader(`
{
 "app_name": "AppNameFromJson",
 "number_of_retries": 5
}
`)

	config := struct {
		AppNameFromJSON             string `json:"app_name" env:"APP_NAME" default:"configurate"`
		NumberOfRetriesInJSONAndEnv int    `json:"number_of_retries" env:"NUMBER_OF_RETRIES"`
		VersionMissing              string `json:"version" default:"1.0default"`
	}{}
	err = LoadAll(&config, NewJSONLoader(json), NewEnvLoader(), NewDefaultsLoader(), NewRequiredLoader())
	if err != nil {
		t.Fatalf("error %s", err.Error())
	}

	if config.AppNameFromJSON != "AppNameFromJson" {
		t.Fatalf("AppNameFromJSON not as specified but: %q", config.AppNameFromJSON)
	}

	if config.NumberOfRetriesInJSONAndEnv != 12 {
		t.Fatalf("NumberOfRetriesInJSONAndEnv not as specified but: %d", config.NumberOfRetriesInJSONAndEnv)
	}

	if config.VersionMissing != "1.0default" {
		t.Fatalf("VersionMissing not as specified but: %q", config.VersionMissing)
	}
}

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

func TestRequiredLoader(t *testing.T) {
	type requiredTestConf struct {
		ReqYes    string
		IntReqYes int
		ReqNo     *string
	}

	conf := requiredTestConf{
		ReqYes:    "present",
		IntReqYes: 1,
	}
	err := NewRequiredLoader().Load(&conf)
	if err != nil {
		t.Fatalf("error %s", err.Error())
	}

	conf = requiredTestConf{
		IntReqYes: 1,
	}
	err = NewRequiredLoader().Load(&conf)
	if err == nil {
		t.Fatalf("error is nil, should be present")
	}
	if err.Error() != "Required Value \"ReqYes\" missing" {
		t.Fatalf("Error message not as expected. Is: %q", err.Error())
	}

	conf = requiredTestConf{
		ReqYes: "present",
	}
	err = NewRequiredLoader().Load(&conf)
	if err == nil {
		t.Fatalf("error is nil, should be present")
	}
	if err.Error() != "Required Value \"IntReqYes\" missing" {
		t.Fatalf("Error message not as expected. Is: %q", err.Error())
	}
}

func TestEnvLoader(t *testing.T) {
	config := struct {
		Unconfigured string
		Used         string `env:"TEST_USED"`
		Number       int    `env:"TEST_INT"`
	}{}

	reverter, err := changeEnv(map[string]string{
		"TEST_USED": "env_value",
		"TEST_INT":  "2",
	})
	if err != nil {
		t.Fatalf("error preparing env %s", err.Error())
	}
	defer reverter()

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
}

func changeEnv(changes map[string]string) (reverter func(), err error) {
	fallbackTo := map[string]string{}

	for key, value := range changes {
		fallbackTo[key] = os.Getenv(key)

		err := os.Setenv(key, value)
		if err != nil {
			return nil, err
		}
	}

	return func() {
		for key, value := range fallbackTo {
			err := os.Setenv(key, value)
			if err != nil {
				panic(err)
			}
		}
	}, nil
}
