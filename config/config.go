package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Server struct {
		Port               string `default:"8080" envconfig:"FLIGHTAGG_PORT"`
		Secret             []byte `required:"true" envconfig:"FLIGHTAGG_SECRET"`
		GoogleClientID     string `required:"true" envconfig:"FLIGHTAGG_GOOGLE_CLIENT_ID"`
		GoogleClientSecret string `required:"true" envconfig:"FLIGHTAGG_GOOGLE_CLIENT_SECRET"`
		GithubClientID     string `required:"true" envconfig:"FLIGHTAGG_GITHUB_CLIENT_ID"`
		GithubClientSecret string `required:"true" envconfig:"FLIGHTAGG_GITHUB_CLIENT_SECRET"`
		Domain             string `envconfig:"FLIGHTAGG_DOMAIN"`
	}
	Database struct {
		MongodbUri string `required:"true" envconfig:"FLIGHTAGG_MONGODB_URI"`
	}
	Email struct {
		SendGridAPIKey string `required:"true" envconfig:"FLIGHTAGG_SENDGRID_API_KEY"`
		FromEmail      string `required:"true" envconfig:"FLIGHTAGG_FROM_EMAIL"`
	}
	UIOrigin string `default:"http://localhost:3000" envconfig:"FLIGHTAGG_UI_ORIGIN"`
}

var Cfg Config

func LoadConfig() {
	b, err := strconv.ParseBool(os.Getenv("PROD"))
	if err == nil && b {
		fmt.Println("In PROD env")
		Cfg = loadConfigFromVariables()
		return
	}

	curDir, err := os.Getwd()
	if err != nil {
		log.Fatal("Cannot get cwd: ", err.Error())
	}
	err = godotenv.Load(curDir + "/.env")
	if err != nil {
		log.Fatal("Cannot load .env")
	}

	var cfg Config
	err = envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal("Cannot parse env variables: ", err.Error())
	}
	Cfg = cfg
}

func loadConfigFromVariables() Config {
	cfg := Config{}
	cfg.Server.Port = os.Getenv("PORT")
	cfg.Server.Secret = []byte(os.Getenv("FLIGHTAGG_SECRET"))
	cfg.Server.GoogleClientID = os.Getenv("FLIGHTAGG_GOOGLE_CLIENT_ID")
	cfg.Server.GoogleClientSecret = os.Getenv("FLIGHTAGG_GOOGLE_CLIENT_SECRET")
	cfg.Server.GithubClientID = os.Getenv("FLIGHTAGG_GITHUB_CLIENT_ID")
	cfg.Server.GithubClientSecret = os.Getenv("FLIGHTAGG_GITHUB_CLIENT_SECRET")
	cfg.Server.Domain = os.Getenv("FLIGHTAGG_DOMAIN")
	cfg.Database.MongodbUri = os.Getenv("FLIGHTAGG_MONGODB_URI")
	cfg.Email.SendGridAPIKey = os.Getenv("FLIGHTAGG_SENDGRID_API_KEY")
	cfg.Email.FromEmail = os.Getenv("FLIGHTAGG_FROM_EMAIL")
	return cfg
}
