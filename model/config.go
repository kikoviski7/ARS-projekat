package model

type Config struct {
	// todo: dodati atribute
	Name    string            `json: "name"`
	Id      string            `json: "id"`
	Params  map[string]string `json: "params"`
	Version int               `json: "version"`
}

type ConfigGroup struct {
	Name    string   `json: "name"`
	Configs []Config `json: "configs"`
	Version int      `json: "version"`
}

// todo: dodati metode

type ConfigRepository interface {
	// todo: dodati metode
	Add(config Config)
	Get(name string, version int) (Config, error)
	GetAll() (map[string]Config, error)
	DeleteByVersion(name string, version int) (Config, error)
	DeleteGroupByVersion(name string, version int) (ConfigGroup, error)
	AddGroup(configGroup ConfigGroup)
}
