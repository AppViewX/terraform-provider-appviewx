package config

type AppViewXEnvironment struct {
	AppViewXUserName             string
	AppViewXPassword             string
	AppViewXEnvironmentIP        string
	AppViewXEnvironmentPort      string
	AppViewXIsHTTPS              bool
	AppViewXClientId             string
	AppViewXClientSecret         string
	ProviderCertDownloadPassword string
	ProviderKeyDownloadPassword  string
	// TfVarsFilePath is the optional path to the .tfvars file that holds appviewx_client_secret.
	// When explicitly set in the provider block, the file is automatically updated after a successful secret regeneration.
	// If empty, regenerated secrets are displayed in logs for manual configuration.
	TfVarsFilePath string
}
