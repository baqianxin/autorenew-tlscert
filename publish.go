package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

type SSLContent struct {
	SNIs          []string `json:"snis"`
	Certificate   string   `json:"cert"`
	Key           string   `json:"key"`
	ValidityStart int64    `json:"validity_start"`
	ValidityEnd   int64    `json:"validity_end"`
}

func getExistingCertificate() string {
	// Get all SSL certificates from APISIX and filter by SNI name
	sniName := getFirstSNINameFromPEM(certPath)
	fmt.Print(sniName)
	reqURL := api_host + apiURL

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		fmt.Println("Error creating HTTP request:", err)
		os.Exit(1)
	}

	req.Header.Set("X-API-KEY", token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error sending HTTP request:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading HTTP response body:", err)
		os.Exit(1)
	}

	var certContent string = string(body)
	return certContent
}

func createNewCertificate() string {
	// Read domains from PEM file by openssl
	snis := getSNIsFromPEM(certPath)

	// Create SSL content JSON
	certContent := SSLContent{
		SNIs:          snis,
		Certificate:   readFile(certPath),
		Key:           readFile(certPath),
		ValidityStart: getValidityTimestamp(certPath, "startdate"),
		ValidityEnd:   getValidityTimestamp(certPath, "enddate"),
	}

	// Convert SSL content struct to JSON
	certJSON, err := json.Marshal(certContent)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		os.Exit(1)
	}

	// POST request to create SSL certificate on APISIX
	reqURL := api_host + apiURL
	reqBody := bytes.NewBuffer(certJSON)
	req, err := http.NewRequest("POST", reqURL, reqBody)
	if err != nil {
		fmt.Println("Error creating HTTP request:", err)
		os.Exit(1)
	}

	req.Header.Set("X-API-KEY", token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error sending HTTP request:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading HTTP response body:", err)
		os.Exit(1)
	}

	return string(respBody)
}

func getFirstSNINameFromPEM(pemPath string) string {
	cmd := exec.Command("openssl", "x509", "-in", pemPath, "-noout", "-text")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error running openssl command:", err)
		os.Exit(1)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "DNS:") {
			sniName := strings.TrimSpace(strings.Split(line, "DNS:")[1])
			return sniName
		}
	}
	return ""
}

type SNIs []string

type Node struct {
	Value SNIs `json:"value"`
}

type NodesResponse struct {
	Node struct {
		Nodes []Node `json:"nodes"`
	} `json:"node"`
}

func getSNIsFromPEM(pemPath string) []string {
	cmd := exec.Command("openssl", "x509", "-in", pemPath, "-noout", "-text")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error running openssl command:", err)
		os.Exit(1)
	}

	var snis []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "DNS:") {
			sni := strings.TrimSpace(strings.Split(line, "DNS:")[1])
			snis = append(snis, sni)
		}
	}
	return snis
}

func getValidityTimestamp(pemPath string, flag string) int64 {
	cmd := exec.Command("openssl", "x509", "-in", pemPath, "-noout", "-"+flag)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error running openssl command:", err)
		os.Exit(1)
	}

	var dateStr string
	if flag == "startdate" {

		dateStr = strings.TrimSpace(strings.ReplaceAll(string(output), "notBefore=", ""))
	}
	if flag == "enddate" {
		dateStr = strings.TrimSpace(strings.ReplaceAll(string(output), "notAfter=", ""))

	}
	dateLayout := "Jan 2 15:04:05 2006 MST" // Adjust date layout as per your output format

	validityTime, err := time.Parse(dateLayout, dateStr)
	if err != nil {
		fmt.Println("Error parsing date:", err)
		os.Exit(1)
	}

	return validityTime.Unix()
}

func readFile(filePath string) string {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading file:", err)
		os.Exit(1)
	}
	return string(content)
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
