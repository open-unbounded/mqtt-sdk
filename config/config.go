package config

type PushConfig struct {
	Brokers            []string
	Username           string `json:",optional"`
	Password           string `json:",optional"`
	ClientIdPrefix     string `json:",default=mqtt-"`
	Conns              int    `json:",default=1"`
	Pushers            int    `json:",default=8"`
	FileStoreDirPrefix string `json:",default=mqtt-data-"`
}
