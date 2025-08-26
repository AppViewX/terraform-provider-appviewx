package main

import (
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"terraform-provider-appviewx/appviewx"
)

var (
	version     = "1.0.7"
	releaseDate = "August 26, 2025"
	description = "Optimization in Download Certificate Resource"
)

func init() {
	log.Println("[INFO] version", version)
	log.Println("[INFO] releaseDate", releaseDate)
	log.Println("[INFO] description", description)
}

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() *schema.Provider {
			return appviewx.Provider()
		},
	})
}
