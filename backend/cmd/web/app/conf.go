package app

const (
	MODE_DEBUG      = "debug"
	MODE_PRODUCTION = "production"
)

type conf struct {
	Mode          string `yaml:"mode"`
	Port          string `yaml:"port"`
	SessionSecret string `yaml:"session_secret"`
	DSN           string `yaml:"dsn"`
	TimezoneLoc   string `yaml:"timezone_loc"`
	FilePath      string `yaml:"file_path"`
	EguneAPI      struct {
		BaseURL string `yaml:"base_url"`
		APIKey  string `yaml:"api_key"`
		Model   string `yaml:"model"`
	} `yaml:"egune_api"`
	OCRAPI struct {
		BaseURL string `yaml:"base_url"`
		APIKey  string `yaml:"api_key"`
	} `yaml:"ocr_api"`
}
