package main

import (
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"terraform-provider-appviewx/appviewx"
)

var (
	version     = "1.0.5"
	releaseDate = "July 15, 2025"
	description = "Create Certificate and Push to AKV, Revoke Certificate in AppViewX, Remove Certificate from AKV and Search Certificate in AppViewX via Custom Workflow."
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
