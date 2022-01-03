package cli

import (
	"encoding/json"
	"fmt"

	"github.com/redhatinsights/insights-operator-conditional-gathering/internal/config"
)

func PrintConfiguration(conf config.Configuration) error {
	// err should be nil as config contains valid fields
	configBytes, err := json.MarshalIndent(conf, "", "    ")
	fmt.Println(string(configBytes))

	return err
}
