package fileops

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
)

func GetFileContentsInMap(fileName string) map[string]interface{} {
	output := make(map[string]interface{})
	log.Println("[DEBUG] fileName : ", fileName)
	if fileName == "" {
		log.Println("[ERROR] File name is empty : ", fileName)
		return output
	}
	masterFile, err := os.Open(fileName)
	if err != nil {
		log.Println("[ERROR] Error in opening the file : ", fileName)
		return output
	}
	masterFileContents, err := ioutil.ReadAll(masterFile)
	if err != nil {
		log.Println("[ERROR] Error in reading the file contents")
	}
	json.Unmarshal(masterFileContents, &output)
	return output
}

func WriteContentsToFile(input map[string]interface{}, outputFileName string) error {
	inputContents, err := json.MarshalIndent(input, "", "    ")
	if err != nil {
		log.Println("[ERROR] Error in Unmarshalling ", err)
		return err
	}

	err = ioutil.WriteFile(outputFileName, inputContents, 0777)
	if err != nil {
		log.Println("[ERROR] Error in Unmarshalling ", err)
		return err
	}
	return nil
}

// escapeForRegexReplacement escapes a string to be used safely in regexp.ReplaceAllString replacement parameter.
// In Go regex replacement strings, $ is special and needs to be escaped as $$ to produce a literal $.
// This function escapes all special characters that have meaning in replacement strings:
// - $ → $$ (produces literal $)
// - \ → \\ (produces literal \)
func escapeForRegexReplacement(s string) string {
	// First escape backslashes, then escape dollar signs
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `$`, `$$`)
	return s
}

// UpdateTfVarsClientSecret updates the client secret in the specified tfvars file if it exists.
// If filePath is empty, displays the secret prominently for manual configuration.
// If filePath is provided but file doesn't exist, returns error and displays secret in logs.
// The file is written with permissions 0600 to protect the secret at rest.
func UpdateTfVarsClientSecret(filePath, clientID, newSecret string) error {
	if filePath == "" {
		// No tfvars file path provided in provider block
		// Display regenerated secret prominently for manual configuration
		displayRegeneratedSecret(clientID, newSecret)
		return fmt.Errorf("no tfvars file configured\n\n=== NEW CLIENT SECRET REGENERATED ===\nClient ID: %s\nClient Secret: %s\n\nPlease configure one of the following:\n1. Add appviewx_tfvars_file_path to your provider block\n2. Update appviewx_client_secret in your .tf file\n3. Update APPVIEWX_TERRAFORM_CLIENT_SECRET environment variable\n\nThen re-run: terraform apply", clientID, newSecret)
	}

	// Check if the file exists
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			log.Printf("[WARN] tfvars file specified in appviewx_tfvars_file_path does not exist: %s", filePath)
			displayRegeneratedSecret(clientID, newSecret)
			return fmt.Errorf("tfvars file not found: %s\n\nNEW CLIENT SECRET:\nClient ID: %s\nClient Secret: %s\n\nUpdate your configuration and re-run terraform apply.", filePath, clientID, newSecret)
		}
		return fmt.Errorf("cannot access tfvars file: %s: %w", filePath, err)
	}

	// File exists, read and update it
	raw, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("could not read tfvars file %s: %w", filePath, err)
	}

	// Match: appviewx_client_secret = "any value" (leading whitespace / spacing variants)
	re := regexp.MustCompile(`(?m)^(\s*appviewx_client_secret\s*=\s*)"[^"]*"`)
	// Escape special characters in the secret for safe use in replacement string.
	// In regex replacement strings, $ must be escaped as $$ to produce a literal $.
	escapedSecret := escapeForRegexReplacement(newSecret)
	updated := re.ReplaceAllString(string(raw), `${1}"`+escapedSecret+`"`)
	if updated == string(raw) {
		log.Printf("[WARN] Key 'appviewx_client_secret' not found in %s", filePath)
		displayRegeneratedSecret(clientID, newSecret)
		return fmt.Errorf("key 'appviewx_client_secret' not found in %s\n\nNEW CLIENT SECRET:\nClient ID: %s\nClient Secret: %s\n\nAdd this to your tfvars file and re-run terraform apply.", filePath, clientID, newSecret)
	}

	if err := os.WriteFile(filePath, []byte(updated), 0600); err != nil {
		return fmt.Errorf("could not write updated secret to %s: %w", filePath, err)
	}

	log.Printf("[INFO] appviewx_client_secret updated successfully in %s", filePath)
	return nil
}

// displayRegeneratedSecret displays the regenerated secret prominently in logs
func displayRegeneratedSecret(clientID, newSecret string) {
	log.Println("")
	log.Println("╔══════════════════════════════════════════════════════════════════════════════╗")
	log.Println("║                    ⚠️  CLIENT SECRET REGENERATED  ⚠️                          ║")
	log.Println("╚══════════════════════════════════════════════════════════════════════════════╝")
	log.Println("")
	log.Printf("Client ID:     %s\n", clientID)
	log.Printf("Client Secret: %s\n", newSecret)
	log.Println("")
	log.Println("Update your configuration and re-run: terraform apply")
	log.Println("")
}
