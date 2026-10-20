package config

// Config agrupa todos los ajustes resueltos al arrancar el servicio.
//
// Se construye una sola vez y se inyecta por parámetro en cada componente, para
// evitar estado global y permitir que cualquier pieza se pruebe con una
// configuración explícita.
type Config struct {
	Port              string
	AppEnv            string
	DBHost            string
	DBPort            int
	DBUser            string
	DBPassword        string
	DBName            string
	DBSSLMode         string
	EvolutionAPIURL   string
	EvolutionAPIKey   string
	EvolutionInstance string
	JIDGrupoPrueba    string
	HooURL            string
}

// SimulationMode indica si el cliente de Evolution debe evitar la red por
// completo. Se activa al no haber URL configurada, y es lo que permite levantar
// el servicio íntegro sin WhatsApp y sin credenciales reales.
func (c *Config) SimulationMode() bool {
	return c.EvolutionAPIURL == ""
}

// Load resuelve la configuración completa desde el entorno, aplicando valores
// por defecto seguros para desarrollo. La instancia canónica es Glauk —el
// TitleCase que usa el panel de Evolution— y puede sobreescribirse con
// EVOLUTION_INSTANCE sin recompilar el binario.
func Load() *Config {
	loadDotEnv()
	return &Config{
		Port:              getEnv("PORT", "3003"),
		AppEnv:            getEnv("APP_ENV", "development"),
		DBHost:            getEnv("DB_HOST", "localhost"),
		DBPort:            getEnvInt("DB_PORT", 5432),
		DBUser:            getEnv("DB_USER", "postgres"),
		DBPassword:        getEnv("DB_PASSWORD", ""),
		DBName:            getEnv("DB_NAME", "automat"),
		DBSSLMode:         getEnv("DB_SSLMODE", "disable"),
		EvolutionAPIURL:   getEnv("EVOLUTION_API_URL", ""),
		EvolutionAPIKey:   getEnv("EVOLUTION_API_KEY", ""),
		EvolutionInstance: getEnv("EVOLUTION_INSTANCE", "Glauk"),
		JIDGrupoPrueba:    getEnv("JID_GRUPO_PRUEBA", ""),
		HooURL:            getEnv("HOO_GLOBAL_RATES_URL", "https://hoo.jairokov.com/global-rates"),
	}
}
