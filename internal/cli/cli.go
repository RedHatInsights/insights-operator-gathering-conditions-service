/*
Copyright © 2021, 2022 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cli

import (
	"encoding/json"
	"fmt"

	"github.com/redhatinsights/insights-operator-gathering-conditions-service/internal/config"
)

const (
	authorsMessage = "Ricardo Lüders, Serhii Zakharov, and CCX Processing team members, Red Hat Inc."
)

// PrintConfiguration function displays actual configuration on standard
// output.
func PrintConfiguration(conf config.Configuration) {
	// err should be nil as config contains valid fields
	configBytes, err := json.MarshalIndent(conf, "", "    ")
	if err != nil {
		// in case of error, don't print the marshalled buffer
		fmt.Printf("error retrieving configuration: %s\n", err)
		return
	}

	fmt.Println(string(configBytes))
}

// PrintAuthors function displays list of authors on standard output.
func PrintAuthors() {
	fmt.Println(authorsMessage)
}
