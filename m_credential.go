package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/uuid"
)

const WavveCredentialPath = "./credentials.json"

// 定义存储数据的结构体
type MCredentials struct {
	GUID     string `json:"guid"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func GetWavveMCredntials() MCredentials {
	var creds MCredentials
	// Try to read the JSON file
	data, err := os.ReadFile(WavveCredentialPath)
	if err != nil {
		if os.IsNotExist(err) {
			LogInfo("File does not exist, creating a new one")
		} else {
			LogError("Failed to read file: ", err)
			return creds
		}
	} else {
		// File exists, parse JSON
		err = json.Unmarshal(data, &creds)
		if err != nil {
			LogError("Failed to parse JSON: ", err)
			//return
		}
	}

	// Check if GUID is empty
	if creds.GUID == "" {
		// Generate a new GUID
		creds.GUID = uuid.New().String()

		// // Set default values if other fields are empty
		// if creds.Username == "" {
		// 	creds.Username = "default_user"
		// }
		// if creds.Password == "" {
		// 	creds.Password = "default_password"
		// }

		// Save the new credentials to file
		err = saveMCredentials(WavveCredentialPath, creds)
		if err != nil {
			LogError("Failed to save file: ", WavveCredentialPath, err)
			return creds
		}
		LogInfo("New credentials created and saved")
	} else {
		LogInfo("Credentials loaded successfully:")
	}

	// Print current credentials
	fmt.Printf("GUID: %s\nUsername: %s\nPassword: %s\n",
		creds.GUID, creds.Username, creds.Password)
	return creds
}

// Save credentials to JSON file
func saveMCredentials(filename string, creds MCredentials) error {
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}
