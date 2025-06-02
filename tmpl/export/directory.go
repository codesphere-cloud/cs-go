package export

import (
	"fmt"
	"os"
)

func CreateDirectory(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Create the directory if it does not exist
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return fmt.Errorf("error creating directory: %v\n", err)
		}

		fmt.Printf("directory %s created successfully\n", path)
	} else {
		fmt.Printf("directory %s already exists\n", path)
	}
	return nil
}
