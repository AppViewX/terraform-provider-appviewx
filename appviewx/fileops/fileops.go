package fileops

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
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

// UpdateTfVarsClientSecret replaces the appviewx_client_secret value in a .tfvars file with
// newSecret.  If filePath is empty the function searches for credentials.auto.tfvars then
// terraform.tfvars in the current working directory.  If no .tfvars file is found but the
// APPVIEWX_TERRAFORM_CLIENT_SECRET environment variable is set, the env var is updated
// in-process so the new secret is used for the remainder of the current terraform run.
// The file is written with permissions 0600 to protect the secret at rest.
func UpdateTfVarsClientSecret(filePath, newSecret string) error {
	if filePath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("could not determine working directory: %w", err)
		}
		candidates := []string{
			filepath.Join(cwd, "credentials.auto.tfvars"),
			filepath.Join(cwd, "terraform.tfvars"),
		}
		for _, c := range candidates {
			if _, statErr := os.Stat(c); statErr == nil {
				filePath = c
				break
			}
		}
		if filePath == "" {
			// No .tfvars file — fall back to updating the env variable if it is set.
			if os.Getenv("APPVIEWX_TERRAFORM_CLIENT_SECRET") != "" {
				// Update the env var for the remainder of this process run.
				if err := os.Setenv("APPVIEWX_TERRAFORM_CLIENT_SECRET", newSecret); err != nil {
					return fmt.Errorf("could not update APPVIEWX_TERRAFORM_CLIENT_SECRET env variable: %w", err)
				}
				// os.Setenv only affects this process; write a sourceable env file so the
				// user can persist the new secret in their shell after terraform exits.
				envFile := filepath.Join(cwd, ".appviewx_secret.env")
				envLine := fmt.Sprintf("export APPVIEWX_TERRAFORM_CLIENT_SECRET='%s'\n", newSecret)
				if writeErr := os.WriteFile(envFile, []byte(envLine), 0600); writeErr != nil {
					log.Printf("[WARN] Could not write env file %s: %v", envFile, writeErr)
				}
				log.Println("[INFO] Client secret was regenerated.")
				log.Printf("[INFO] The new secret has been written to: %s", envFile)
				log.Println("[INFO] NOTE: os.Setenv only affects the current process. Run the following")
				log.Println("[INFO] command in your shell to persist it for future terraform runs:")
				log.Printf("[INFO]   source %s", envFile)
				log.Printf("[INFO]   # or:  export APPVIEWX_TERRAFORM_CLIENT_SECRET='%s'", newSecret)
				return nil
			}
			// Neither a .tfvars file nor an env variable is available.
			// Log the new secret clearly so the user can update their provider configuration manually.
			log.Println("[WARN] ============================================================")
			log.Println("[WARN] Client secret was regenerated but no persistence target found.")
			log.Println("[WARN] New client secret: " + newSecret)
			log.Println("[WARN] Please update 'appviewx_client_secret' in your provider")
			log.Println("[WARN] configuration or .tfvars file with the value above, then")
			log.Println("[WARN] re-run `terraform apply` to avoid re-triggering regeneration.")
			log.Println("[WARN] ============================================================")
			return nil
		}
	}

	raw, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("could not read tfvars file %s: %w", filePath, err)
	}

	// Match:  appviewx_client_secret = "any value"  (leading whitespace / spacing variants)
	re := regexp.MustCompile(`(?m)^(\s*appviewx_client_secret\s*=\s*)"[^"]*"`)
	updated := re.ReplaceAllString(string(raw), `${1}"`+newSecret+`"`)
	if updated == string(raw) {
		return fmt.Errorf("key 'appviewx_client_secret' not found in %s; cannot update secret", filePath)
	}

	if err := os.WriteFile(filePath, []byte(updated), 0600); err != nil {
		return fmt.Errorf("could not write updated secret to %s: %w", filePath, err)
	}

	log.Printf("[INFO] appviewx_client_secret updated successfully in %s", filePath)
	return nil
}
