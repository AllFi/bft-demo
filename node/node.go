package node

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/spf13/viper"
	abci "github.com/tendermint/tendermint/abci/types"
	cfg "github.com/tendermint/tendermint/config"
	tmflags "github.com/tendermint/tendermint/libs/cli/flags"
	"github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"
	nm "github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/proxy"
	"github.com/tendermint/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

func InitNewNode(path string, nodeIndex int) (id string, address string, err error) {
	config := cfg.DefaultConfig()
	config = config.SetRoot(path + "/.tendermint")
	cfg.EnsureRoot(path + "/.tendermint")
	config.LogLevel = "none"
	config.Consensus.WalPath = path + "/.tendermint/data/cs.wal/wal"

	initFilesWithConfig(config)
	if err != nil {
		panic(err)
	}

	nodeKey, err := p2p.LoadNodeKey(config.NodeKeyFile())
	if err != nil {
		return
	}

	id = string(nodeKey.ID())
	address = "127.0.0.1:" + strconv.Itoa(shiftPort(26656, nodeIndex))
	return
}

func initFilesWithConfig(config *cfg.Config) error {
	// private validator
	privValKeyFile := config.PrivValidatorKeyFile()
	privValStateFile := config.PrivValidatorStateFile()
	var pv *privval.FilePV
	if tmos.FileExists(privValKeyFile) {
		pv = privval.LoadFilePV(privValKeyFile, privValStateFile)
	} else {
		pv = privval.GenFilePV(privValKeyFile, privValStateFile)
		pv.Save()
	}

	nodeKeyFile := config.NodeKeyFile()
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	if tmos.FileExists(nodeKeyFile) {
		logger.Info("Found node key", "path", nodeKeyFile)
	} else {
		if _, err := p2p.LoadOrGenNodeKey(nodeKeyFile); err != nil {
			return err
		}
		logger.Info("Generated node key", "path", nodeKeyFile)
	}

	// genesis file
	genFile := config.GenesisFile()
	if tmos.FileExists(genFile) {
		logger.Info("Found genesis file", "path", genFile)
	} else {
		genDoc := types.GenesisDoc{
			ChainID:         fmt.Sprintf("test-chain"),
			GenesisTime:     tmtime.Now(),
			ConsensusParams: types.DefaultConsensusParams(),
		}
		pubKey, err := pv.GetPubKey()
		if err != nil {
			return fmt.Errorf("can't get pubkey: %w", err)
		}
		genDoc.Validators = []types.GenesisValidator{{
			Address: pubKey.Address(),
			PubKey:  pubKey,
			Power:   10,
		}}

		if err := genDoc.SaveAs(genFile); err != nil {
			return err
		}
		logger.Info("Generated genesis file", "path", genFile)
	}

	return nil
}

func RunNewNode(app abci.Application, path string, nodeIndex int, persistentPeers string) (err error) {
	configFile := path + "/.tendermint/config/config.toml"
	// flag.Lookup("config").Value.Set(configFile)

	node, err := newTendermint(app, configFile, nodeIndex, persistentPeers)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(2)
	}

	node.Start()
	defer func() {
		node.Stop()
		node.Wait()
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	os.Exit(0)

	return
}

func newTendermint(app abci.Application, configFile string, nodeIndex int, persistentPeers string) (*nm.Node, error) {
	// read config
	config := cfg.DefaultConfig()
	config.RootDir = filepath.Dir(filepath.Dir(configFile))
	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("viper failed to read config file: %w", err)
	}
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("viper failed to unmarshal config: %w", err)
	}
	if err := config.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("config is invalid: %w", err)
	}

	// create logger
	logger := log.NewTMLogger(log.NewSyncWriter(ioutil.Discard))
	//logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	var err error
	logger, err = tmflags.ParseLogLevel(config.LogLevel, logger, "none")
	if err != nil {
		return nil, fmt.Errorf("failed to parse log level: %w", err)
	}

	// read private validator
	pv := privval.LoadFilePV(
		config.PrivValidatorKeyFile(),
		config.PrivValidatorStateFile(),
	)

	// read node key
	nodeKey, err := p2p.LoadNodeKey(config.NodeKeyFile())
	if err != nil {
		return nil, fmt.Errorf("failed to load node's key: %w", err)
	}

	config.ProxyApp = "tcp://127.0.0.1:" + strconv.Itoa(shiftPort(26658, nodeIndex))              //tcp://127.0.0.1:26658
	config.Instrumentation.PrometheusListenAddr = ":" + strconv.Itoa(shiftPort(26660, nodeIndex)) //:26660
	config.P2P.ListenAddress = "tcp://0.0.0.0:" + strconv.Itoa(shiftPort(26656, nodeIndex))       //tcp://0.0.0.0:26656
	config.RPC.ListenAddress = "tcp://127.0.0.1:" + strconv.Itoa(shiftPort(26657, nodeIndex))     //tcp://127.0.0.1:26657
	config.P2P.PersistentPeers = persistentPeers
	config.Consensus.WalPath = filepath.Dir(filepath.Dir(configFile)) + "/data/cs.wal/wal"

	// create node
	node, err := nm.NewNode(
		config,
		pv,
		nodeKey,
		proxy.NewLocalClientCreator(app),
		nm.DefaultGenesisDocProviderFunc(config),
		nm.DefaultDBProvider,
		nm.DefaultMetricsProvider(config.Instrumentation),
		logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create new Tendermint node: %w", err)
	}

	fmt.Println("Started: " + strconv.Itoa(nodeIndex))
	return node, nil
}

func shiftPort(basePort int, index int) int {
	return basePort + 10*index
}
