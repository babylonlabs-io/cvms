package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// root config
type MonitoringConfig struct {
	Monikers     []string      `yaml:"monikers"`
	ChainConfigs []ChainConfig `yaml:"chains"`
}

// each chain
type ChainConfig struct {
	DisplayName       string         `yaml:"display_name"`
	ChainID           string         `yaml:"chain_id"`
	TrackingAddresses []string       `yaml:"tracking_addresses,omitempty"`
	Nodes             []NodeEndPoint `yaml:"nodes"`
	ProviderNodes     []NodeEndPoint `yaml:"provider_nodes"`
	Monikers          []string       `yaml:"monikers,omitempty"`
}

// each chain's available node list
type NodeEndPoint struct {
	RPC  string `yaml:"rpc"`
	API  string `yaml:"api"`
	GRPC string `yaml:"grpc"`
}

func replacePlaceholders(byteString []byte) []byte {
	// Simple function to replace placeholders in the config file with their corresponding environment variables.
	// Variables take the form of ${VAR_NAME}

	// Convert the byte slice to a string
	cfg := string(byteString)

	// Regular expression to match placeholders in the format ${VAR_NAME}
	re := regexp.MustCompile(`\$\{([a-zA-Z_][a-zA-Z0-9_]*)\}`)

	cfg = re.ReplaceAllStringFunc(cfg, func(match string) string {
		varName := re.FindStringSubmatch(match)[1]
		if value, exists := os.LookupEnv(varName); exists {
			return value
		}
		return match // return original placeholder if not found in environment variables
	})
	return []byte(cfg)
}

// TODO: ignore failed chains
func GetConfig(path string) (*MonitoringConfig, error) {
	dataBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read config file")
	}

	dataBytes = replacePlaceholders(dataBytes)

	cfg := &MonitoringConfig{}
	err = yaml.Unmarshal(dataBytes, cfg)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode config file")
	}

	_, err = validateChainName(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to validate your config")
	}

	return cfg, nil
}

func validateChainName(cfg *MonitoringConfig) (bool, error) {
	supportChains, err := GetSupportChainConfig()
	if err != nil {
		return false, errors.Wrap(err, "failed to get support chain config")
	}
	_ = supportChains

	// Validate chains in config
	for _, cc := range cfg.ChainConfigs {
		_, exist := supportChains.Chains[cc.ChainID]
		if !exist {
			return false, errors.Errorf("config has unsupported name in your config: %s.\nCheck your chain-id the error target. chain-id(%s) would be matched in one of support_chains[./docker/cvms/support_chains.yaml]", cc.DisplayName, cc.ChainID)
		}
	}

	return true, nil
}

// GetSupportChain WorkingDirectory in anywhere
func MustGetSupportChainPath(configName string) string {
	wd, _ := os.Getwd()

	var rootPath string
	for _, path := range strings.SplitAfter(wd, "cvms") {
		rootPath = path
		break
	}

	return fmt.Sprintf("%s/docker/cvms/%s", rootPath, configName)
}
