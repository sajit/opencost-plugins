package harness

import (
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/opencost/opencost/core/pkg/model"
	ocplugin "github.com/opencost/opencost/core/pkg/plugin"
)

// the test harness is designed to run plugins locally, and return the results
// the harness expects to be given a path to a valid config, and a path to a plugin implementation
// it does not run binaries, and instead uses go run
func InvokePlugin(pathToConfig, pathToPluginSrc string, req model.CustomCostRequest) []model.CustomCostResponse {
	filename := path.Base(pathToConfig)
	pluginName := strings.Split(filename, "_")[0]
	// Create an hclog.Logger
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "plugin",
		Output: os.Stdout,
		Level:  hclog.Debug,
	})
	var handshakeConfig = plugin.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "PLUGIN_NAME",
		MagicCookieValue: pluginName,
	}
	// pluginMap is the map of plugins we can dispense.
	var pluginMap = map[string]plugin.Plugin{
		"CustomCostSource": &ocplugin.CustomCostPlugin{},
	}
	// We're a host! Start by launching the plugin process.
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
		Cmd:             exec.Command("go", "run", pathToPluginSrc, pathToConfig),
		Logger:          logger,
	})

	defer client.Kill()

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		log.Fatal(err)
	}

	// Request the plugin
	raw, err := rpcClient.Dispense("CustomCostSource")
	if err != nil {
		log.Fatal(err)
	}

	src := raw.(ocplugin.CustomCostSource)
	var iface model.CustomCostRequestInterface = &req
	resp := src.GetCustomCosts(iface)
	return resp
}
