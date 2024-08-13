package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

func ReadFile(filename string) []string {
	content, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	// Split the content by lines and store it in an array
	lines := strings.Split(string(content), "\n")

	// Filter out empty lines
	var nonEmptyLines []string
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine != "" {
			nonEmptyLines = append(nonEmptyLines, trimmedLine)
		}
	}

	return nonEmptyLines
}

func ReadStdIn() []string {

	var inputLines []string
	var content []byte

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		inputLines = append(inputLines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	content = []byte(strings.Join(inputLines, "\n"))

	lines := strings.Split(string(content), "\n")

	// Filter out empty lines
	var nonEmptyLines []string
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine != "" {
			nonEmptyLines = append(nonEmptyLines, trimmedLine)
		}
	}

	return nonEmptyLines
}

func CommandParser() (string, string) {
	// Define flags
	var targetsFile string
	var payloadsFile string

	flag.StringVar(&targetsFile, "t", "", "Targets File")
	flag.StringVar(&payloadsFile, "p", "", "Payloads File")

	// Parse the flags
	flag.Parse()

	return targetsFile, payloadsFile
}

func Scanning(targets []string, payloads []string) {
	for i := 0; i < len(targets); i++ {
		sanitizedURL := strings.TrimSpace(targets[i])
		u, err := url.Parse(sanitizedURL)
		if err != nil {
			panic(err)
		}

		// Parse the query parameters
		m, _ := url.ParseQuery(u.RawQuery)

		// Store the default values
		defaults := make(map[string]string)
		for key, values := range m {
			if len(values) > 0 {
				defaults[key] = values[0]
			}
		}

		// Loop through the map, set each parameter with a payload, and reset it
		for key := range m {
			for j := 0; j < len(payloads); j++ {
				// Set the parameter with a payload without encoding
				m[key] = []string{strings.Join(m[key], " ") + payloads[j]}
				newParams := ""

				// Reconstruct the raw query and print the modified URL
				u.RawQuery = m.Encode()
				for key, vals := range u.Query() {
					for _, val := range vals {
						if newParams != "" {
							newParams += "&"
						}
						newParams += fmt.Sprintf("%s=%s", key, val)
					}
				}
				// raw query without encoding
				u.RawQuery = newParams

				PostRequest(u.String(), payloads[j])

				// Reset the parameter to its default value
				m[key] = []string{defaults[key]}
			}
		}
	}
}

func PostRequest(url string, grepPattern string) {
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	// Grep a pattern from the response
	re := regexp.MustCompile(grepPattern)
	matches := re.FindAllString(string(body), -1)

	// Check if the pattern is found
	found := len(matches) > 0
	printVuln(url, grepPattern)

	if found {
		printVuln(url, grepPattern)
	}
}

func printVuln(url string, payload string) {
	fmt.Println(url, "["+payload+"]")
}

func main() {
	targetsFile, payloadsFile := CommandParser()

	var targets []string

	switch targetsFile {
	case "":
		targets = ReadStdIn()
	default:
		targets = ReadFile(targetsFile)
	}

	if len(targets) == 0 || payloadsFile == "" {
		fmt.Println("Error: Both -t (Targets File) and -p (Payloads File) flags are required.")
		flag.Usage()
		os.Exit(1)
	}

	payloads := ReadFile(payloadsFile)

	Scanning(targets, payloads)
}
