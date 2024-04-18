package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

const (
	apiURL = "/apisix/admin/ssls"
)

var (
	acme     string
	certPath string
	api_host string
	token    string
	domain   string
	force    bool
	debug    bool
)

var renewCmd = &cobra.Command{
	Use:   "renew",
	Short: "Renew SSL certificate and upload to APISIX",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Renewing SSL certificate and uploading to APISIX...")
		run(cmd)
	},
}

func init() {
	// Add flags for certificate renewal options

	// Params for certificate renewal:
	//
	// - `domain`: the domain for which the certificate is being renewed
	// - `path`: the path to store certificate files
	// - `control_host`: the APISIX control console host
	// - `token`: the APISIX control console token
	// - `force`: renew the cert if certificate files exist
	renewCmd.PersistentFlags().StringP("acme", "a", "~/.acme.sh/acme.sh", "Path of acme.sh")
	renewCmd.PersistentFlags().StringP("path", "p", ".", "Path to store certificate files")
	renewCmd.PersistentFlags().StringP("control_host", "x", "", "APISIX control console host")
	renewCmd.PersistentFlags().StringP("token", "t", "", "APISIX control console token")
	renewCmd.PersistentFlags().StringP("domain", "d", "", "Domain for certificate renewal")
	renewCmd.PersistentFlags().StringP("dns_channel", "c", "", "DNS channel for certificate renewal")
	renewCmd.PersistentFlags().BoolP("debug", "", false, "acme.sh debug mode")
	// Add other flags as needed
	renewCmd.Flags().BoolP("force", "f", false, "Renew the cert if certificate files exist")

}

func run(cmd *cobra.Command) {
	// create params interface for acme, certPath, api_host, token, domain, force
	acme, _ = cmd.PersistentFlags().GetString("acme")
	certPath, _ = cmd.PersistentFlags().GetString("path")
	api_host, _ = cmd.PersistentFlags().GetString("control_host")
	token, _ = cmd.PersistentFlags().GetString("token")
	domain, _ = cmd.PersistentFlags().GetString("domain")
	force, _ = cmd.Flags().GetBool("force")
	debug, _ = cmd.Flags().GetBool("debug")

	fmt.Println(acme, certPath, api_host, token, domain, force, debug)
	// check params
	checkParams(acme, certPath, api_host, token, domain, force, debug)

	// install acme.sh
	InstallAcmeSh(acme)
	if RenewCert(acme, certPath, api_host, token, domain, force, debug) {
		fmt.Println("Certificate renewed and uploaded to APISIX")
		err := PublishCert(certPath, api_host, token, domain, force, debug)
		if err != nil {
			fmt.Println(err)
		}
	}

}

func checkParams(acme, certPath, api_host, token, domain string, force, debug bool) {
	if isValidDomain(domain) {
		return
	}
	fmt.Println("Invalid domain")
	os.Exit(1)
}

func isValidDomain(domain string) bool {
	// 域名不能为空
	if domain == "" {
		return false
	}

	// 域名长度不能超过255个字符
	if len(domain) > 255 {
		return false
	}

	// 检查域名中每个标签的合法性
	labels := strings.Split(domain, ".")

	if len(labels) <= 1 {
		return false
	}

	if labels[0] == "*" {
		labels = labels[1:]
	}

	for _, label := range labels {

		if len(label) > 63 || len(label) == 0 {
			return false
		}
		if !isValidLabel(label) {
			return false
		}
	}

	return true
}

func isValidLabel(label string) bool {
	// 使用正则表达式检查标签是否合法
	// 标签必须由字母、数字和连字符组成，且连字符不能在开头或结尾
	regex := regexp.MustCompile(`^[a-zA-Z0-9](?:[a-zA-Z0-9-]*[a-zA-Z0-9])?$`)
	return regex.MatchString(label)
}

func main() {
	// Execute the root command
	if err := renewCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
