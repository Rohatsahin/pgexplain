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
	"html/template"
	"os"
	"path/filepath"
)

// This plan template was provided by pev2 visualization library
// https://github.com/dalibo/pev2?tab=readme-ov-file#without-building-tools
const planTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{ .Title }}</title>
    <script src="https://unpkg.com/vue@3.2.45/dist/vue.global.prod.js"></script>
    <script src="https://unpkg.com/pev2/dist/pev2.umd.js"></script>
    <link href="https://unpkg.com/bootstrap@5.3.2/dist/css/bootstrap.min.css" rel="stylesheet" />
    <link rel="stylesheet" href="https://unpkg.com/pev2/dist/style.css" />
</head>
<body>
    <div id="app">
        <pev2 :plan-source="plan" :plan-query="query" />
    </div>
    <script>
        const { createApp } = Vue;

        const plan = "{{ .Plan }}"
        const query = "{{ .Query }}"

        const app = createApp({
            data() {
                return {
                    plan: plan,
                    query: query
                };
            },
        });
        app.component("pev2", pev2.Plan);
        app.mount("#app");
    </script>
</body>
</html>
`

type TemplateData struct {
	Title string
	Plan  string
	Query string
}

// writePlan generates an HTML file with the execution plan and query.
// It returns the file absolute path of the generated file.
func writePlan(plan, query, title string) string {
	name := title + ".html"
	data := TemplateData{
		Title: title,
		Plan:  plan,
		Query: query,
	}

	// Parse and execute the template
	tmpl, err := template.New("plan").Parse(planTemplate)
	if err != nil {
		logErrorAndExit("unable to parse plan template: ", err)
	}

	// Output to a file
	file, err := os.Create(name)
	if err != nil {
		logErrorAndExit("unable to create plan file: ", err)
	}
	defer file.Close()

	// Execute the template with data
	err = tmpl.Execute(file, data)
	if err != nil {
		logErrorAndExit("unable to render plan template: ", err)
	}

	// Get the absolute path of the created file
	abs, err := filepath.Abs(file.Name())
	if err != nil {
		logErrorAndExit("unable get plan file absolute path: ", err)
	}

	return abs
}
