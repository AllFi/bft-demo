package node

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/spf13/viper"
	abci "github.com/tendermint/tendermint/abci/types"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/bytes"
	tmflags "github.com/tendermint/tendermint/libs/cli/flags"
	"github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	nm "github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/proxy"
	"github.com/tendermint/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

func InitNewNodes(basePath string, count int) (persistentPeers string, err error) {
	config := cfg.DefaultConfig()
	genVals := make([]types.GenesisValidator, count)

	peers := make([]string, 0)
	for i := 0; i < count; i++ {
		nodeDir := basePath + "/node" + strconv.Itoa(i) + "/.tendermint"
		config = config.SetRoot(nodeDir)
		cfg.EnsureRoot(nodeDir)
		config.Consensus.WalPath = nodeDir + "/data/cs.wal/wal"

		if err := initFilesWithConfig(config); err != nil {
			return "", err
		}

		pvKeyFile := filepath.Join(nodeDir, config.BaseConfig.PrivValidatorKey)
		pvStateFile := filepath.Join(nodeDir, config.BaseConfig.PrivValidatorState)
		pv := privval.LoadFilePV(pvKeyFile, pvStateFile)

		pubKey, err := pv.GetPubKey()
		if err != nil {
			return "", fmt.Errorf("can't get pubkey: %w", err)
		}
		genVals[i] = types.GenesisValidator{
			Address: pubKey.Address(),
			PubKey:  pubKey,
			Power:   1,
			Name:    "node" + strconv.Itoa(i),
		}

		nodeKey, err := p2p.LoadNodeKey(config.NodeKeyFile())
		if err != nil {
			return "", err
		}

		id := string(nodeKey.ID())
		address := "127.0.0.1:" + strconv.Itoa(ShiftPort(26656, i))
		peers = append(peers, id+"@"+address)
	}

	persistentPeers = strings.Join(peers, ",")

	// Generate genesis doc from generated validators
	genDoc := &types.GenesisDoc{
		ChainID:         "chain-" + tmrand.Str(6),
		ConsensusParams: types.DefaultConsensusParams(),
		GenesisTime:     tmtime.Now(),
		InitialHeight:   0,
		Validators:      genVals,
	}

	// Write genesis file.
	for i := 0; i < count; i++ {
		nodeDir := basePath + "/node" + strconv.Itoa(i) + "/.tendermint"
		if err := genDoc.SaveAs(filepath.Join(nodeDir, config.BaseConfig.Genesis)); err != nil {
			_ = os.RemoveAll(basePath)
			return "", err
		}
	}

	// Overwrite default config.
	for i := 0; i < count; i++ {
		nodeDir := basePath + "/node" + strconv.Itoa(i) + "/.tendermint"
		config.SetRoot(nodeDir)
		config.P2P.AddrBookStrict = false
		config.P2P.AllowDuplicateIP = true
		config.P2P.PersistentPeers = persistentPeers
		config.Moniker = bytes.HexBytes(tmrand.Bytes(8)).String()

		cfg.WriteConfigFile(filepath.Join(nodeDir, "config", "config.toml"), config)
	}

	fmt.Printf("Successfully initialized %v node directories\n", count)
	return persistentPeers, nil
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

	return nil
}

func Run(app abci.Application, basePath string, nodeIndex int) (err error) {
	configFile := basePath + "/node" + strconv.Itoa(nodeIndex) + "/.tendermint/config/config.toml"

	node, err := newTendermint(app, configFile, nodeIndex)
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

func newTendermint(app abci.Application, configFile string, nodeIndex int) (*nm.Node, error) {
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
	//logger := log.NewTMLogger(log.NewSyncWriter(ioutil.Discard))
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	var err error
	logger, err = tmflags.ParseLogLevel("error", logger, "error")
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

	config.ProxyApp = "tcp://127.0.0.1:" + strconv.Itoa(ShiftPort(26658, nodeIndex)) //tcp://127.0.0.1:26658
	config.Instrumentation.Prometheus = true
	config.Instrumentation.PrometheusListenAddr = ":" + strconv.Itoa(ShiftPort(26660, nodeIndex)) //:26660
	config.P2P.ListenAddress = "tcp://0.0.0.0:" + strconv.Itoa(ShiftPort(26656, nodeIndex))       //tcp://0.0.0.0:26656
	config.RPC.ListenAddress = "tcp://127.0.0.1:" + strconv.Itoa(ShiftPort(26657, nodeIndex))     //tcp://127.0.0.1:26657
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

	fmt.Println("Started node with index: " + strconv.Itoa(nodeIndex))
	fmt.Println("RPC listen address: " + config.RPC.ListenAddress)
	return node, nil
}

func ShiftPort(basePort int, index int) int {
	return basePort + 10*index
}
