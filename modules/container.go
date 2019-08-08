package modules

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sarulabs/di"
	"github.com/spf13/viper"
)

const (
	ConfigPath        = "CONFIG_PATH"
	DefaultConfigPath = "./configs"
	InstDatabase      = "primary_db"
	InstAppConfig     = "primary_config"

	ConfigKeyDbHost     = "app.database.host"
	ConfigKeyDbUser     = "app.database.user"
	ConfigKeyDbPassword = "app.database.password"
	ConfigKeyDbName     = "app.database.name"
	ConfigKeyDbPort     = "app.database.port"
)

// InitContainer - Initialize the container and bootstrap the resources
func InitContainer() (container di.Container, err error) {
	var (
		builder *di.Builder
	)

	if builder, err = di.NewBuilder(); err != nil {
		err = errors.New("failed to initialize application container")
		return
	}

	if err = InitAppConfig(builder); err != nil {
		return
	}

	if err = InitDatabase(builder); err != nil {
		return
	}

	// build container
	container = builder.Build()

	if container == nil {
		err = errors.New("failed to initialize application container")
		return
	}

	return
}

// InitAppConfig - Initialize application config and store in container
func InitAppConfig(builder *di.Builder) (err error) {

	err = builder.Add(
		di.Def{
			Name:  InstAppConfig,
			Scope: di.App,
			Build: func(ctn di.Container) (i interface{}, e error) {
				path := os.Getenv(ConfigPath)
				c := viper.New()
				c.SetConfigName("config")
				c.SetConfigType("yaml")
				if path == "" {
					c.AddConfigPath(DefaultConfigPath) // look for config in the working directory
					log.Println("Loading configs from default location...")
				} else {
					c.AddConfigPath(path)
					log.Printf("Loading configs from location: %s\n", path)
				}

				err = c.ReadInConfig()
				if err != nil {
					return
				}
				i = c
				return
			},
		})

	return
}

// InitDatabase - Initialize database and store in container
// This retunr the default sql database connection which can further wrapper and used to support multiple database
func InitDatabase(builder *di.Builder) (err error) {

	err = builder.Add(
		di.Def{
			Name:  InstDatabase,
			Scope: di.App,
			Build: func(ctn di.Container) (i interface{}, e error) {

				var db *sql.DB
				// get database configs
				host := ctn.Get(InstAppConfig).(*viper.Viper).Get(ConfigKeyDbHost).(string)
				user := ctn.Get(InstAppConfig).(*viper.Viper).Get(ConfigKeyDbUser).(string)
				pass := ctn.Get(InstAppConfig).(*viper.Viper).Get(ConfigKeyDbPassword).(string)
				name := ctn.Get(InstAppConfig).(*viper.Viper).Get(ConfigKeyDbName).(string)
				port := ctn.Get(InstAppConfig).(*viper.Viper).Get(ConfigKeyDbPort).(int)

				dbSource := user + ":" + pass + "@tcp(" + host + ":" + fmt.Sprint(port) + ")/" + name
				db, e = sql.Open("mysql", dbSource)

				if e != nil {
					return
				}

				return db, nil
			},
			Close: func(obj interface{}) error {
				return obj.(*sql.DB).Close()
			},
		})

	return
}
