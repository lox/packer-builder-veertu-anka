package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/plugin"
	packerVersion "github.com/hashicorp/packer-plugin-sdk/version"
	"github.com/veertuinc/packer-plugin-veertu-anka/builder/anka"
	"github.com/veertuinc/packer-plugin-veertu-anka/post-processor/ankaregistry"
)

var (
	version = ""
	commit  = ""
)

func main() {
	pps := plugin.NewSet()
	pps.RegisterBuilder("vm-create", new(anka.Builder))
	pps.RegisterBuilder("vm-clone", new(anka.Builder))
	pps.RegisterPostProcessor("registry-push", new(ankaregistry.PostProcessor))
	var pluginVersion = packerVersion.InitializePluginVersion(version, commit)
	pps.SetVersion(pluginVersion)
	log.Printf("plugin version: %s+%s", pluginVersion.GetVersion(), pluginVersion.GetVersionPrerelease())
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
