package model

type Config struct {
	// todo: dodati atribute
	Name    string            `json: "name"`
	Version int               `json: "version"`
	Params  map[string]string `json: "params"`
}

// todo: dodati metode

type ConfigRepository interface {
	// todo: dodati metode
	Add(config Config)
	Get(name string, version int) (Config, error)
}
