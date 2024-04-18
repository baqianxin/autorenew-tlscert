# autorenew-tlscert

使用 `acme.sh` 新签/续签 Lets Encrypt 免费 ssl 证书，并通过 APISIX Control API 发布更新
默认域名验证方式使用 cloudflare dns 服务：需要设置 环境变量 CF_key CF_email

## usage

```bash
go run . renew -c "http://xxx.apisix.qianxin.me:9180" -d "*.sec.qianxin.me" -t"xxxxxtoken" --path /Users/qianxinqba/.acme.sh/*.sec.qianxin.me_ecc  --force true --debug true

qianxinqba@CN_LF7CQ3N4XG autorenew-tlscert %  go run . renew -h
qianxinqba@CN_LF7CQ3N4XG autorenew-tlscert % ./goacme renew -h
Renew SSL certificate and upload to APISIX

Usage:
  renew [flags]

Flags:
  -a, --acme string           Path of acme.sh (default "~/.acme.sh/acme.sh")
  -x, --control_host string   APISIX control console host
  -c, --dns_channel string    DNS channel for certificate renewal(default "dns_cf")
  -d, --domain string         Domain for certificate renewal
  -f, --force                 Renew the cert if certificate files exist
  -h, --help                  help for renew
  -p, --path string           Path to store certificate files (default ".")
  -t, --token string          APISIX control console token
      --debug boolean          acme.sh debug mode (default false) 

```
