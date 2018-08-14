package configurate

import (
	"reflect"
	"encoding/json"
	"io"
	"errors"
	"fmt"
	"strconv"
	"os"
)

type Loader interface {
	Load(target interface{}) error
}

func LoadAll(target interface{}, loaders ...Loader) error {
	for _, loader := range loaders {
		if err := loader.Load(target); err != nil {
			return err
		}
	}
	return nil
}

type JSONLoader struct {
	Reader io.Reader
}

func NewJSONLoader(reader io.Reader) *JSONLoader {
	return &JSONLoader{Reader: reader}
}

func (l *JSONLoader) Load(target interface{}) error {
	decoder := json.NewDecoder(l.Reader)
	return decoder.Decode(target)
}

type DefaultsLoader struct {
}

func NewDefaultsLoader() *DefaultsLoader {
	return &DefaultsLoader{}
}

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
			return errors.New(fmt.Sprintf("No default loader defined for type %s", field.Type))
		}
	}

	return nil
}

type EnvLoader struct {
}

func NewEnvLoader() *EnvLoader {
	return &EnvLoader{}
}

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
			return errors.New(fmt.Sprintf("No env loader defined for type %s", field.Type))
		}
	}

	return nil
}
