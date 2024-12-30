/*
Package cmd

# Copyright © 2024 Rohat Sahin

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
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// these remote services are provided by Dalibo, and thanks to them this service
// https://github.com/dalibo/pev2?tab=readme-ov-file#dalibo-service-recommended
const (
	uploadURL = "https://explain.dalibo.com/new.json"
	accessURL = "https://explain.dalibo.com/plan/%s"
)

type UploadResponse struct {
	ID        string `json:"id"`
	DeleteKey string `json:"deleteKey"`
}

// uploadPlan uploads a query execution plan and returns the access URL.
func uploadPlan(plan, query, title string) string {
	formData := url.Values{
		"plan":  {plan},
		"query": {query},
		"title": {title},
	}

	// post form for plan
	response, err := http.PostForm(uploadURL, formData)
	if err != nil {
		logErrorAndExit("failed to upload plan to remote: ", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		logErrorAndExit("failed to read upload body result: ", err)
	}

	// parse json response to struct
	var uploadResponse UploadResponse
	err = json.Unmarshal(body, &uploadResponse)
	if err != nil {
		logErrorAndExit("failed to parse response: ", err)
	}

	return fmt.Sprintf(accessURL, uploadResponse.ID)
}
