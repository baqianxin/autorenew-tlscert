package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"
)

// It uses the acme.sh tool to automatically renew certificates for a domain
// and publishes the certificate files through an HTTP API to the APISIX control console.

// Step 1: Install acme.sh if not present
func InstallAcmeSh(acme string) {
	if !checkAcmeShPresence(acme) &&
		downloadAndInstall(acme) {
		return
	}
}

func checkAcmeShPresence(acme string) bool {
	// Define the command to check if acme.sh is installed
	fmt.Println(acme, "-v")
	cmd := exec.Command("sh", "-c", acme+" -v")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Run the command
	if err := cmd.Run(); err != nil {
		// acme.sh is not installed or not found in PATH
		fmt.Println("acme.sh is not installed or not found in PATH:", err)
		return false
	}

	// acme.sh is installed
	fmt.Println("acme.sh is installed")
	return true
}

func downloadAndInstall(acme string) bool {
	// Define the command to download and install acme.sh
	cmd := exec.Command("curl", "https://get.acme.sh", "-o", "./install-acme.sh")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	err := cmd.Run()
	if err != nil {
		// Error occurred during download and installation
		fmt.Println("Error downloading and installing acme.sh:", err)
		return false
	}

	cmd = exec.Command("bash", "./install-acme.sh")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		// Error occurred during download and installation
		fmt.Println("Error downloading and installing acme.sh:", err)
		return false
	}

	// Print the output of the command (optional)
	return true
}

// step 2: renew certificate
func RenewCert(acme, certPath, api_host, token, domain string, force, debug bool) bool {
	// Define the command to renew the certificate using acme.sh
	script := acme + " --issue  -d \"" + domain + "\" --dns dns_cf "
	if certPath != "" {
		script = script + " -w \"" + certPath + "\""
	}

	if debug {
		script = script + " --staging --debug"
	}
	if force {
		script = script + " --force"
	}
	print(script)
	cmd := exec.Command("sh", "-c", script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Run the command
	err := cmd.Run()
	if err != nil {
		// Error occurred during certificate renewal
		fmt.Println("Error renewing certificate:", err)
		return false
	}

	// Print the output of the command (optional)
	return true

}

type SSLListContent struct {
	Value struct {
		ID            string   `json:"id"`
		SNIs          []string `json:"snis"`
		Certificate   string   `json:"cert"`
		Key           string   `json:"key"`
		ValidityStart int64    `json:"validity_start"`
		ValidityEnd   int64    `json:"validity_end"`
	} `json:"value"`
}

type SSLPostContent struct {
	SNIs          []string `json:"snis"`
	Certificate   string   `json:"cert"`
	Key           string   `json:"key"`
	ValidityStart int64    `json:"validity_start,omitempty"`
	ValidityEnd   int64    `json:"validity_end,omitempty"`
}
type SSLList struct {
	List  []SSLListContent `json:"list"`
	Total int              `json:"total"`
}

// step 3: publish certificate

// step 5: Check if the certificate already exists in the APISIX SSL List
// 请求APISIX SSL 接口返回数据格式如下 ：
//
//	{
//	    "total": 1,
//	    "list": [
//	        {
//	            "createdIndex": 383691,
//	            "value": {
//	                "snis": [
//	                    "*.home.qianxin.me"
//	                ],
//	                "update_time": 1713594846,
//	                "status": 1,
//	                "validity_end": 1721139220,
//	                "validity_start": 1713363221,
//	                "type": "server",
//	                "id": "0000000000199211",
//	                "create_time": 1713594842,
//	                "cert": "xxxxxxx",
//	                "key": "xxxxxxx"
//	            },
//	            "modifiedIndex": 383692,
//	            "key": "/apisix/ssls/0000000000199211"
//	        }
//	    ]
//	}
func PublishCert(certPath, api_host, token, domain string, force, debug bool) error {
	// step 4: Check if the certificate files already exists in the certPath
	certFile := certPath + "/" + domain + ".cer"
	keyFile := certPath + "/" + domain + ".key"
	_, err := os.Stat(certFile)
	if err != nil {
		fmt.Println("Certificate files :" + certFile + " not exists, use --force to renew")
		return err
	}

	fmt.Println("Certificate files: " + certFile + " already exists, use --force to renew")
	reqURL := api_host + "/apisix/admin/ssls"

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-API-KEY", token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Error sending HTTP request: %v\n", err)
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode == 200 {
		fmt.Println("ok")
	} else {
		fmt.Println(string(body))
		os.Exit(1)
	}

	var sslList SSLList
	err = json.Unmarshal(body, &sslList)
	if err != nil {
		return err
	}
	fmt.Println(sslList.List[0].Value.SNIs)
	updateId := ""
	for i := 0; i < len(sslList.List); i++ {
		if domain == sslList.List[i].Value.SNIs[0] {
			fmt.Println(sslList.List[i].Value.SNIs)
			updateId = sslList.List[i].Value.ID
			break

		}
		fmt.Println(sslList.List[i].Value.SNIs)
	}

	certContent := new(SSLPostContent)
	certContent.SNIs = []string{domain}
	certContent.Certificate = readFile(certFile)
	certContent.Key = readFile(keyFile)
	certContent.ValidityStart = getValidityTimestamp(certFile, "startdate")
	certContent.ValidityEnd = getValidityTimestamp(certFile, "enddate")

	if updateId == "" && force {
		updateId = "199200" + strconv.FormatInt(time.Now().Unix(), 10)
	}
	UpdateSSLData(updateId, api_host, token, certContent)

	return nil
}

func UpdateSSLData(id, api_host, token string, certContent *SSLPostContent) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}
	reqURL := api_host + "/apisix/admin/ssls" + "/" + id
	fmt.Println(reqURL)
	// Convert SSL content struct to JSON
	certJSON, err := json.Marshal(certContent)
	fmt.Println(string(certJSON))
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return err
	}

	// POST request to create SSL certificate on APISIX
	reqBody := bytes.NewBuffer(certJSON)
	req, err := http.NewRequest("PUT", reqURL, reqBody)
	if err != nil {
		fmt.Println("Error creating HTTP request:", err)
		return err
	}

	req.Header.Set("X-API-KEY", token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error sending HTTP request:", err)
		return err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading HTTP response body:", err)
		return err

	}

	fmt.Println(string(respBody))
	return nil
}
