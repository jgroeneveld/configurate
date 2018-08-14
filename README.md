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

err := configurate.LoadFile("config.json", &config)
```

## Adding sources / loaders

configurate uses a `Loader` interface and the `LoadAll` method to be extendable.

Just make sure the order of the loaders makes sense.

```
type Loader interface {
    Load(target interface{}) error
}

err := configurate.LoadAll(&target, loader1, loader2)
```

Interesting loaders would be more formats or a consul extension.