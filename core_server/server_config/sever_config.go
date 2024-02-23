package server_config
import (
	"fmt"
	"os"
	"github.com/spf13/viper"
)

const (
	SERVICE_PORT = 1234
	CONFIG_PORT = 1235
	
)

type ServerConfig struct {
	PrimaryNode int
	NodePort int
	BackUpNode []int
	NeedSync bool
	
}

func (cfg *ServerConfig) PingToPrimaryNode(current_port int, reply *bool) error {
	*reply = true
	return nil
}

func (cfg *ServerConfig) RegisterBackUpNode(port int, reply  *bool) error {
	fmt.Println("Registering BackUp Node: ", port)
	cfg.BackUpNode = append(cfg.BackUpNode, port)
	fmt.Println("BackUp Node: ", cfg.BackUpNode)
	cfg.WriteConfig()
	*reply = true
	cfg.NeedSync = true
	return nil
}

func ReadConfig() (*ServerConfig, error){
	cfg := new(ServerConfig)
	viper.SetConfigName("server_config.yaml")
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("Error: ", err)
		return cfg, err
	}
	cfg = viper.Get("config").(*ServerConfig)
	return cfg, nil
}

func (cfg *ServerConfig) WriteConfig(){
	viper.Set("config", cfg)
	viper.SetConfigType("yaml") 
	cwd, _ := os.Getwd()
	confg_path := cwd + "/server_config.yaml"
	viper.WriteConfigAs(confg_path)
}

func (cfg *ServerConfig) IsPrimary(port int, reply *bool) error {
	if cfg.PrimaryNode == port {
		*reply = true
	} else {
		*reply = false
	}
	return nil
}

func (cfg *ServerConfig) UpdateAndWriteConfig(new_cfg ServerConfig, reply *bool) error{
	fmt.Println("[UpdateAndWriteConfig]:: Updating config at ",cfg.NodePort)
	cfg.PrimaryNode = new_cfg.PrimaryNode
	cfg.BackUpNode = new_cfg.BackUpNode
	cfg.NeedSync = false
	viper.Set("config", cfg)
	viper.SetConfigType("yaml") 
	cwd, _ := os.Getwd()
	confg_path := cwd + "/server_config.yaml"
	viper.WriteConfigAs(confg_path)
	*reply = true
	fmt.Println(cfg)
	return nil
}