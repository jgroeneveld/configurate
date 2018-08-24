package configurate

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
)

// LoadFile loads a configuration from file into target and uses the default loaders.
func LoadFile(path string, target interface{}) error {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return err
	}

	return LoadAll(target, NewJSONLoader(file), NewEnvLoader(), NewDefaultsLoader(), NewRequiredLoader())
}

// Loader is an interface to define generic configuration loaders
type Loader interface {
	Load(target interface{}) error
}

// LoadAll runs all loaders and stops when an error occurs
func LoadAll(target interface{}, loaders ...Loader) error {
	for _, loader := range loaders {
		if err := loader.Load(target); err != nil {
			return err
		}
	}
	return nil
}

// JSONLoader is a Loader for JSON
type JSONLoader struct {
	Reader io.Reader
}

// NewJSONLoader returns a JSONLoader
func NewJSONLoader(reader io.Reader) *JSONLoader {
	return &JSONLoader{Reader: reader}
}

// Load executes the loading
func (l *JSONLoader) Load(target interface{}) error {
	decoder := json.NewDecoder(l.Reader)
	err := decoder.Decode(target)
	if err != nil {
		return errors.New("JSONError: " + err.Error())
	}
	return nil
}

// DefaultsLoader is a loader that loads the defaults from the `default` tag, if there is not already a value
type DefaultsLoader struct {
}

// NewDefaultsLoader returns a DefaultsLoader
func NewDefaultsLoader() *DefaultsLoader {
	return &DefaultsLoader{}
}

// Load executes the loading
func (l *DefaultsLoader) Load(target interface{}) error {
	targetValue := reflect.ValueOf(target).Elem()
	targetType := targetValue.Type()

	for i := 0; i < targetType.NumField(); i++ {
		field := targetType.Field(i)
		tag := field.Tag.Get("default")
		if tag == "" {
			continue
		}

		f := targetValue.FieldByName(field.Name)
		switch field.Type.Kind() {
		case reflect.String:
			if f.String() == "" {
				f.SetString(tag)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if f.Int() == 0 {
				converted, err := strconv.Atoi(tag)
				if err != nil {
					return err
				}
				f.SetInt(int64(converted))
			}
		default:
			return fmt.Errorf("No default loader defined for type %s", field.Type)
		}
	}

	return nil
}

// EnvLoader loads from the env, specified by the env tag
type EnvLoader struct {
}

// NewEnvLoader returns a EnvLoader
func NewEnvLoader() *EnvLoader {
	return &EnvLoader{}
}

// Load executes the loading
func (l *EnvLoader) Load(target interface{}) error {
	targetValue := reflect.ValueOf(target).Elem()
	targetType := targetValue.Type()

	for i := 0; i < targetType.NumField(); i++ {
		field := targetType.Field(i)
		tag := field.Tag.Get("env")
		if tag == "" {
			continue
		}

		value := os.Getenv(tag)
		if value == "" {
			continue
		}

		f := targetValue.FieldByName(field.Name)
		switch field.Type.Kind() {
		case reflect.String:
			f.SetString(value)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			converted, err := strconv.Atoi(value)
			if err != nil {
				return err
			}
			f.SetInt(int64(converted))
		default:
			return fmt.Errorf("No env loader defined for type %s", field.Type)
		}
	}

	return nil
}

// RequiredLoader checks all non-pointer values for presence and returns an error if not given.
type RequiredLoader struct {
}

// NewRequiredLoader returns a RequiredLoader
func NewRequiredLoader() *RequiredLoader {
	return &RequiredLoader{}
}

// Load executes the loading
func (l *RequiredLoader) Load(target interface{}) error {
	targetValue := reflect.ValueOf(target).Elem()
	targetType := targetValue.Type()

	for i := 0; i < targetType.NumField(); i++ {
		field := targetType.Field(i)

		f := targetValue.FieldByName(field.Name)
		switch field.Type.Kind() {
		case reflect.String:
			if f.String() == "" {
				return fmt.Errorf("Required Value %q missing", field.Name)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if f.Int() == 0 {
				return fmt.Errorf("Required Value %q missing", field.Name)
			}
		case reflect.Ptr:
			// ignore, pointers can be optional
		default:
			return fmt.Errorf("No required loader defined for type %s", field.Type)
		}
	}

	return nil
}
