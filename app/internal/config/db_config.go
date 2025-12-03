package config

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func NewDBConfigFromEnv() (DBConfig, error) {
	host, err := getRequiredString("AUTH_SERVICE_DB_HOST")
	if err != nil {
		return DBConfig{}, err
	}

	port, err := getRequiredString("AUTH_SERVICE_DB_PORT")
	if err != nil {
		return DBConfig{}, err
	}

	user, err := getRequiredString("AUTH_SERVICE_DB_USER")
	if err != nil {
		return DBConfig{}, err
	}

	password, err := getRequiredString("AUTH_SERVICE_DB_PASSWORD")
	if err != nil {
		return DBConfig{}, err
	}

	dbName, err := getRequiredString("AUTH_SERVICE_DB_NAME")
	if err != nil {
		return DBConfig{}, err
	}

	return DBConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbName,
	}, nil
}
