package configs

import (
	"fmt"
	"log"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type EnvConfig struct {
	DBhost      string
	DBport      string
	DBname      string
	DBuser      string
	Dbpasswd    string
	LeveldbPath string
	FabricPath  string
}

var Env EnvConfig

func init() {
	/**  config comand line flags **/
	pflag.String("env", "devp", "the environment to run:  devp/staging/container")
	pflag.String("configpath", "./configs/env-config.yaml", "the path to the config file")
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	/**  config environment variables, take effect if command line flags empty **/
	viper.AutomaticEnv()
	viper.SetEnvPrefix("platform") // will be uppercased automatically
	viper.BindEnv("env")
	viper.BindEnv("configpath")

	env := viper.GetString("env")
	path := viper.GetString("configpath")
	fmt.Printf(" ....... Runtime is setting to %s with path=%s\n", env, path)

	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
	fmt.Printf("Using config: %s\n", viper.ConfigFileUsed())

	viper.Sub(env).Unmarshal(&Env)
	fmt.Printf("--->>> %+v\n", Env)

}
