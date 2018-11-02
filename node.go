package shell

import (
	"context"
	"errors"
	_ "expvar"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path"
	"sync"
	"time"

	"github.com/ipfs/go-ipfs/assets"
	utilmain "github.com/ipfs/go-ipfs/cmd/ipfs/util"
	oldcmds "github.com/ipfs/go-ipfs/commands"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/core/corehttp"
	"github.com/ipfs/go-ipfs/namesys"
	"github.com/ipfs/go-ipfs/repo/fsrepo"
	migrate "github.com/ipfs/go-ipfs/repo/fsrepo/migrations"

	config "gx/ipfs/QmPEpj17FDRpc7K1aArKZp3RsHtzRMKykeK9GVgn4WQGPR/go-ipfs-config"
	ma "gx/ipfs/QmT4U94DnD8FRfqr21obWY32HLM5VExccPKMjQHofeYqr9/go-multiaddr"
	"gx/ipfs/QmYYv3QFnfQbiwmi1tpkgKF8o4xFnZoBrvpupTiGJwL9nH/client_golang/prometheus"
	manet "gx/ipfs/Qmaabb1tJZ2CX5cp6MuuiGgns71NYoxdgQP6Xdid1dVceC/go-multiaddr-net"
	mprome "gx/ipfs/QmbqQP7GMwJCePfEvbMSZmyPifAz1kKZifo8vwVnVHt1wL/go-metrics-prometheus"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var (
	errRepoExists = errors.New("ipfs configuration file already exists")

	// DefaultNodeConfig default node config
	DefaultNodeConfig = NodeConfig{
		Root:       "ipfs",
		StorageMax: "10G",
	}
)

const (
	nBitsForKeypairDefault = 2048

	routingOptionDHTClientKwd = "dhtclient"
	routingOptionDHTKwd       = "dht"
	routingOptionNoneKwd      = "none"
	routingOptionDefaultKwd   = "default"
)

// Logger log
type Logger interface {
	Info(msg string, ctx ...interface{})
	Warn(msg string, ctx ...interface{})
	Error(msg string, ctx ...interface{})
}

// NodeConfig node config
type NodeConfig struct {
	Root   string
	Logger Logger

	// TODO: more config
	StorageMax string // in B, kB, kiB, MB, ...
}

// Node ipfs node
type Node struct {
	cfg      *NodeConfig
	repoRoot string
	logger   Logger
}

// NewNode return a ipfs node
func NewNode(cfg *NodeConfig) *Node {
	return &Node{
		cfg:      cfg,
		repoRoot: cfg.Root,
		logger:   cfg.Logger,
	}
}

// IsInitialized return repo root is initialized
func (n *Node) IsInitialized() bool {
	return fsrepo.IsInitialized(n.repoRoot)
}

// Init implement ipfs init
func (n *Node) Init() error {
	daemonLocked, err := fsrepo.LockedByOtherProcess(n.repoRoot)
	if err != nil {
		return err
	}

	if daemonLocked {
		return fmt.Errorf("ipfs daemon is running")
	}

	if err := checkWritable(n.repoRoot); err != nil {
		return err
	}

	if fsrepo.IsInitialized(n.repoRoot) {
		return errRepoExists
	}

	var conf *config.Config
	conf, err = config.Init(ioutil.Discard, nBitsForKeypairDefault)
	if err != nil {
		return err
	}
	setupNodeConfig(conf, n.cfg)

	if err := fsrepo.Init(n.repoRoot, conf); err != nil {
		return err
	}

	if err := addDefaultAssets(n.repoRoot); err != nil {
		return err
	}

	return initializeIpnsKeyspace(n.repoRoot)
}

func setupNodeConfig(ipfsConf *config.Config, cfg *NodeConfig) {
	if cfg.StorageMax != "" {
		ipfsConf.Datastore.StorageMax = cfg.StorageMax
	}
}

func checkWritable(dir string) error {
	_, err := os.Stat(dir)
	if err == nil {
		// dir exists, make sure we can write to it
		testfile := path.Join(dir, "test")
		fi, err := os.Create(testfile)
		if err != nil {
			if os.IsPermission(err) {
				return fmt.Errorf("%s is not writeable by the current user", dir)
			}
			return fmt.Errorf("unexpected error while checking writeablility of repo root: %s", err)
		}
		fi.Close()
		return os.Remove(testfile)
	}

	if os.IsNotExist(err) {
		// dir doesn't exist, check that we can create it
		return os.Mkdir(dir, 0775)
	}

	if os.IsPermission(err) {
		return fmt.Errorf("cannot write to %s, incorrect permissions", err)
	}

	return err
}

func addDefaultAssets(repoRoot string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r, err := fsrepo.Open(repoRoot)
	if err != nil { // NB: repo is owned by the node
		return err
	}

	nd, err := core.NewNode(ctx, &core.BuildCfg{Repo: r})
	if err != nil {
		return err
	}
	defer nd.Close()

	_, err = assets.SeedInitDocs(nd)
	if err != nil {
		return fmt.Errorf("init: seeding init docs failed: %s", err)
	}

	return nil
}

func initializeIpnsKeyspace(repoRoot string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r, err := fsrepo.Open(repoRoot)
	if err != nil { // NB: repo is owned by the node
		return err
	}

	nd, err := core.NewNode(ctx, &core.BuildCfg{Repo: r})
	if err != nil {
		return err
	}
	defer nd.Close()

	err = nd.SetupOfflineRouting()
	if err != nil {
		return err
	}

	return namesys.InitializeKeyspace(ctx, nd.Namesys, nd.Pinning, nd.PrivateKey)
}

// Daemon implement ipfs daemon
func (n *Node) Daemon(ctx context.Context, ready func()) error {
	// Inject metrics before we do anything
	err := mprome.Inject()
	if err != nil {
		return fmt.Errorf("injecting prometheus handler for metrics failed with message: %s", err.Error())
	}

	_, _, err = utilmain.ManageFdLimit()
	if err != nil {
		return fmt.Errorf("setting file descriptor limit: %s", err.Error())
	}

	cctx := &oldcmds.Context{
		ConfigRoot: n.repoRoot,
		LoadConfig: fsrepo.ConfigAt,
		ReqLog:     &oldcmds.ReqLog{},
	}

	// acquire the repo lock _before_ constructing a node. we need to make
	// sure we are permitted to access the resources (datastore, etc.)
	repo, err := fsrepo.Open(cctx.ConfigRoot)
	switch err {
	default:
		return err
	case fsrepo.ErrNeedMigration:
		err = migrate.RunMigration(fsrepo.RepoVersion)
		if err != nil {
			return err
		}

		repo, err = fsrepo.Open(cctx.ConfigRoot)
		if err != nil {
			return err
		}
	case nil:
		break
	}

	cfg, err := cctx.GetConfig()
	if err != nil {
		return err
	}

	// Start assembling node config
	ncfg := &core.BuildCfg{
		Repo:                        repo,
		Permanent:                   true, // It is temporary way to signify that node is permanent
		Online:                      true,
		DisableEncryptedConnections: false,
		ExtraOpts: map[string]bool{
			"pubsub": true,
			"ipnsps": true,
			"mplex":  true,
		},
	}

	rcfg, err := repo.Config()
	if err != nil {
		return err
	}

	routingOption := rcfg.Routing.Type
	if routingOption == "" {
		routingOption = routingOptionDHTKwd
	}
	switch routingOption {
	case routingOptionDHTClientKwd:
		ncfg.Routing = core.DHTClientOption
	case routingOptionDHTKwd:
		ncfg.Routing = core.DHTOption
	case routingOptionNoneKwd:
		ncfg.Routing = core.NilRouterOption
	default:
		return fmt.Errorf("unrecognized routing option: %s", routingOption)
	}

	node, err := core.NewNode(ctx, ncfg)
	if err != nil {
		return fmt.Errorf("error from node construction: %s", err.Error())
	}
	node.SetLocal(false)

	defer func() {
		// We wait for the node to close first, as the node has children
		// that it will wait for before closing, such as the API server.
		node.Close()

		select {
		case <-ctx.Done():
			n.logger.Info("Gracefully shut down daemon")
		default:
		}
	}()

	cctx.ConstructNode = func() (*core.IpfsNode, error) {
		return node, nil
	}

	// construct api endpoint - every time
	apiErrc, err := serveHTTPApi(cctx, n.logger)
	if err != nil {
		return err
	}

	// construct http gateway - if it is set in the config
	var gwErrc <-chan error
	if len(cfg.Addresses.Gateway) > 0 {
		var err error
		gwErrc, err = serveHTTPGateway(cctx, n.logger)
		if err != nil {
			return err
		}
	}

	// initialize metrics collector
	prometheus.MustRegister(&corehttp.IpfsNodeCollector{Node: node})

	// daemon is ready
	ready()

	// collect long-running errors and block for shutdown
	// TODO(cryptix): our fuse currently doesnt follow this pattern for graceful shutdown
	for err := range merge(apiErrc, gwErrc) {
		if err != nil {
			return err
		}
	}

	return nil
}

// serveHTTPApi collects options, creates listener, prints status message and starts serving requests
func serveHTTPApi(cctx *oldcmds.Context, logger Logger) (<-chan error, error) {
	cfg, err := cctx.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("serveHTTPApi: GetConfig() failed: %s", err)
	}

	apiAddrs := cfg.Addresses.API
	listeners := make([]manet.Listener, 0, len(apiAddrs))
	for _, addr := range apiAddrs {
		apiMaddr, err := ma.NewMultiaddr(addr)
		if err != nil {
			return nil, fmt.Errorf("serveHTTPApi: invalid API address, (err: %s)", err.Error())
		}

		apiLis, err := manet.Listen(apiMaddr)
		if err != nil {
			return nil, fmt.Errorf("serveHTTPApi: manet.Listen(%s) failed: %s", apiMaddr, err)
		}

		// we might have listened to /tcp/0 - lets see what we are listing on
		apiMaddr = apiLis.Multiaddr()
		logger.Info("API server listening on ", apiMaddr)

		listeners = append(listeners, apiLis)
	}

	// by default, we don't let you load arbitrary ipfs objects through the api,
	// because this would open up the api to scripting vulnerabilities.
	// only the webui objects are allowed.
	gatewayOpt := corehttp.GatewayOption(false, corehttp.WebUIPaths...)

	var opts = []corehttp.ServeOption{
		corehttp.MetricsCollectionOption("api"),
		corehttp.CheckVersionOption(),
		corehttp.CommandsOption(*cctx),
		corehttp.WebUIOption,
		gatewayOpt,
		corehttp.VersionOption(),
		defaultMux("/debug/vars"),
		defaultMux("/debug/pprof/"),
		corehttp.MutexFractionOption("/debug/pprof-mutex/"),
		corehttp.MetricsScrapingOption("/debug/metrics/prometheus"),
		corehttp.LogOption(),
	}

	if len(cfg.Gateway.RootRedirect) > 0 {
		opts = append(opts, corehttp.RedirectOption("", cfg.Gateway.RootRedirect))
	}

	node, err := cctx.ConstructNode()
	if err != nil {
		return nil, fmt.Errorf("serveHTTPApi: ConstructNode() failed: %s", err)
	}

	if err := node.Repo.SetAPIAddr(listeners[0].Multiaddr()); err != nil {
		return nil, fmt.Errorf("serveHTTPApi: SetAPIAddr() failed: %s", err)
	}

	errc := make(chan error)
	var wg sync.WaitGroup
	for _, apiLis := range listeners {
		wg.Add(1)
		go func(lis manet.Listener) {
			defer wg.Done()
			errc <- corehttp.Serve(node, manet.NetListener(lis), opts...)
		}(apiLis)
	}

	go func() {
		wg.Wait()
		close(errc)
	}()

	return errc, nil
}

// serveHTTPGateway collects options, creates listener, prints status message and starts serving requests
func serveHTTPGateway(cctx *oldcmds.Context, logger Logger) (<-chan error, error) {
	cfg, err := cctx.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("serveHTTPGateway: GetConfig() failed: %s", err)
	}

	writable := cfg.Gateway.Writable
	gatewayAddrs := cfg.Addresses.Gateway
	listeners := make([]manet.Listener, 0, len(gatewayAddrs))
	for _, addr := range gatewayAddrs {
		gatewayMaddr, err := ma.NewMultiaddr(addr)
		if err != nil {
			return nil, fmt.Errorf("serveHTTPGateway: invalid gateway address: %q (err: %s)", addr, err)
		}

		gwLis, err := manet.Listen(gatewayMaddr)
		if err != nil {
			return nil, fmt.Errorf("serveHTTPGateway: manet.Listen(%s) failed: %s", gatewayMaddr, err)
		}
		// we might have listened to /tcp/0 - lets see what we are listing on
		gatewayMaddr = gwLis.Multiaddr()

		listeners = append(listeners, gwLis)
	}

	var opts = []corehttp.ServeOption{
		corehttp.MetricsCollectionOption("gateway"),
		corehttp.IPNSHostnameOption(),
		corehttp.GatewayOption(writable, "/ipfs", "/ipns"),
		corehttp.VersionOption(),
		corehttp.CheckVersionOption(),
		corehttp.CommandsROOption(*cctx),
	}

	if len(cfg.Gateway.RootRedirect) > 0 {
		opts = append(opts, corehttp.RedirectOption("", cfg.Gateway.RootRedirect))
	}

	node, err := cctx.ConstructNode()
	if err != nil {
		return nil, fmt.Errorf("serveHTTPGateway: ConstructNode() failed: %s", err)
	}

	errc := make(chan error)
	var wg sync.WaitGroup
	for _, lis := range listeners {
		wg.Add(1)
		go func(lis manet.Listener) {
			defer wg.Done()
			errc <- corehttp.Serve(node, manet.NetListener(lis), opts...)
		}(lis)
	}

	go func() {
		wg.Wait()
		close(errc)
	}()

	return errc, nil
}

// merge does fan-in of multiple read-only error channels
// taken from http://blog.golang.org/pipelines
func merge(cs ...<-chan error) <-chan error {
	var wg sync.WaitGroup
	out := make(chan error)

	// Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(c <-chan error) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	for _, c := range cs {
		if c != nil {
			wg.Add(1)
			go output(c)
		}
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

// defaultMux tells mux to serve path using the default muxer. This is
// mostly useful to hook up things that register in the default muxer,
// and don't provide a convenient http.Handler entry point, such as
// expvar and http/pprof.
func defaultMux(path string) corehttp.ServeOption {
	return func(node *core.IpfsNode, _ net.Listener, mux *http.ServeMux) (*http.ServeMux, error) {
		mux.Handle(path, http.DefaultServeMux)
		return mux, nil
	}
}
