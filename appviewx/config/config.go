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
	// TfVarsFilePath is the path to the .tfvars file that holds appviewx_client_secret.
	// When set, the file is automatically updated after a successful secret regeneration.
	TfVarsFilePath string
}
