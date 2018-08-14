# jgroeneveld/configurate


configurate is a simple configuration library.
It can load from JSON files, OS Environment and accepts defaults.

## Example:

```
config := struct {
    AppName         string `json:"app_name"`
    NumberOfRetries int    `json:"number_of_retries" env:"NUMBER_OF_RETRIES"`
    Version         string `json:"version" default:"1.0"`
}{}

err := LoadFile("config.json", &config)
```
