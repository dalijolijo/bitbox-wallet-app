// Copyright 2018 Shift Devices AG
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/digitalbitbox/bitbox-wallet-app/util/errp"
	"github.com/digitalbitbox/bitbox-wallet-app/util/locker"
	"github.com/digitalbitbox/bitbox-wallet-app/util/rpc"
)

// btcCoinConfig holds configurations specific to a btc-based coin.
type btcCoinConfig struct {
	ElectrumServers []*rpc.ServerInfo `json:"electrumServers"`
}

// ethCoinConfig holds configurations for ethereum coins.
type ethCoinConfig struct {
	NodeURL string `json:"nodeURL"`
}

// Backend holds the backend specific configuration.
type Backend struct {
	BitcoinP2PKHActive       bool `json:"bitcoinP2PKHActive"`
	BitcoinP2WPKHP2SHActive  bool `json:"bitcoinP2WPKHP2SHActive"`
	BitcoinP2WPKHActive      bool `json:"bitcoinP2WPKHActive"`
	LitecoinP2WPKHP2SHActive bool `json:"litecoinP2WPKHP2SHActive"`
	LitecoinP2WPKHActive     bool `json:"litecoinP2WPKHActive"`
	BitcoreP2PKHActive       bool `json:"bitcoreP2PKHActive"`
        BitcoreP2WPKHP2SHActive  bool `json:"bitcoreP2WPKHP2SHActive"`
        BitcoreP2WPKHActive      bool `json:"bitcoreP2WPKHActive"`
	EthereumActive           bool `json:"ethereumActive"`

	BTC  btcCoinConfig `json:"btc"`
	TBTC btcCoinConfig `json:"tbtc"`
	LTC  btcCoinConfig `json:"ltc"`
	TLTC btcCoinConfig `json:"tltc"`
	BTX  btcCoinConfig `json:"btx"`
        TBTX btcCoinConfig `json:"tbtx"`
	ETH  ethCoinConfig `json:"eth"`
	TETH ethCoinConfig `json:"teth"`
	RETH ethCoinConfig `json:"reth"`
}

// AccountActive returns the Active setting for a coin by code.
func (backend Backend) AccountActive(code string) bool {
	switch code {
	case "tbtc-p2pkh", "btc-p2pkh", "rbtc-p2pkh":
		return backend.BitcoinP2PKHActive
	case "tbtc-p2wpkh-p2sh", "btc-p2wpkh-p2sh", "rbtc-p2wpkh-p2sh":
		return backend.BitcoinP2WPKHP2SHActive
	case "tbtc-p2wpkh", "btc-p2wpkh", "rbtc-p2wpkh":
		return backend.BitcoinP2WPKHActive
	case "tltc-p2wpkh-p2sh", "ltc-p2wpkh-p2sh":
		return backend.LitecoinP2WPKHP2SHActive
	case "tltc-p2wpkh", "ltc-p2wpkh":
		return backend.LitecoinP2WPKHActive
	case "tbtx-p2pkh", "btx-p2pkh":
                return backend.BitcoreP2PKHActive
        case "tbtx-p2wpkh-p2sh", "btx-p2wpkh-p2sh":
                return backend.BitcoreP2WPKHP2SHActive
        case "tbtx-p2wpkh", "btx-p2wpkh":
                return backend.BitcoreP2WPKHActive
	case "eth", "teth", "reth":
		return backend.EthereumActive
	default:
		panic(fmt.Sprintf("unknown code %s", code))
	}
}

// AppConfig holds the whole app configuration.
type AppConfig struct {
	Backend  Backend     `json:"backend"`
	Frontend interface{} `json:"frontend"`
}

const shiftRootCA = `
-----BEGIN CERTIFICATE-----
MIIGGjCCBAKgAwIBAgIJAKRWPF0NRtHyMA0GCSqGSIb3DQEBDQUAMIGZMQswCQYD
VQQGEwJDSDEPMA0GA1UECAwGWnVyaWNoMR0wGwYDVQQKDBRTaGlmdCBDcnlwdG9z
ZWN1cml0eTEzMDEGA1UECwwqU2hpZnQgQ3J5cHRvc2VjdXJpdHkgQ2VydGlmaWNh
dGUgQXV0aG9yaXR5MSUwIwYDVQQDDBxTaGlmdCBDcnlwdG9zZWN1cml0eSBSb290
IENBMB4XDTE4MDYwODE0NTA0MloXDTM4MDYwMzE0NTA0MlowgZkxCzAJBgNVBAYT
AkNIMQ8wDQYDVQQIDAZadXJpY2gxHTAbBgNVBAoMFFNoaWZ0IENyeXB0b3NlY3Vy
aXR5MTMwMQYDVQQLDCpTaGlmdCBDcnlwdG9zZWN1cml0eSBDZXJ0aWZpY2F0ZSBB
dXRob3JpdHkxJTAjBgNVBAMMHFNoaWZ0IENyeXB0b3NlY3VyaXR5IFJvb3QgQ0Ew
ggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIKAoICAQC0H598q5C+yDJI9F8QYkYK
6/48kFNQ0rbAKcKkgR0+H8CGuFVOGQdcv7tObCMe0Dyr8ioNkq7AP+Nt1e1TVgKQ
ANmJqz2rKvA4sIIgdBjUs0DXPuCaDzGGbJHIXnGMuGANX6xnqvdOj7kIA6r6s7Hh
eWQEB8tGiRdHWJitpkc1xEfW1DhnMQPnSihSJM5qltXVPKxzqqElv0iGI/La3S8W
nJV7kTGTsLouX1CcwLjp6avlVy56utOYRXkgfuY88XxmOjlAECeoYCWFBGaSWK+h
2sBLbRC9G0YWmNCqB+GjMj8myj06crLn7mZgBODEyUrFYMjAPrpmAScmw38y2rwN
AK6ii75P+sHc3BPi05Vap2GoTAY0db62NiN3dsNxHB5DbehA4Zfaqzcakjv4CSRo
zkg2JSlofOZWd3aomxIKfFLl+aVFjukXEKaz8P+2xe5/2/M35kKIIJCuHz1Ybor1
Ze9YmLAnLnbTCA7VcKkUs25lskL/zRC4sdLzgJ2V2UdHWPAo/ttwBXtw6piw4v4N
DfCuKDMiomxwNiGvb3GZWMhOHT30NLZ0nuRAGjeg7jFBqSh2SPeDu+hnImAAh2WX
7ul3/kschLF+3otC/x7jmAMWXzb1oVWRqj2Gyjner1p82gWxj5k7Hs/2mG3TV6sv
pyVqMqonbirxuO+kbzYLxQIDAQABo2MwYTAdBgNVHQ4EFgQU301oVCni/CbDXJ9V
Fez6Vgu0arUwHwYDVR0jBBgwFoAU301oVCni/CbDXJ9VFez6Vgu0arUwDwYDVR0T
AQH/BAUwAwEB/zAOBgNVHQ8BAf8EBAMCAYYwDQYJKoZIhvcNAQENBQADggIBAD1d
/KJ3w1Je3oOx0afcXOf2IOoMvKSFbBg9u+rpXBh60cacjPtwbIMIyF3ynYYGzx7D
x8mr6wagJ+uqKn9E7JGp2h0lKhT9cgxzqIk3r4D/jhvh1zijInCEbPPphwbemzIG
JxxmpDOeURHVxCcSpIJlGfRURdfdXwleWiz9zCkNUvmgTDrfBjEk6ywSSKD4uJuT
jBcav1P4OkeFokAPO1Uc9NCXox5NUAsDosdZUbxbH8vf61Xbr6fnxmy731s9D7cc
djXPb3pbXtRL4A0hNnOWcuPM30hn4ZkIm08TGT+IMOFYBk+pe2IXSFzUcDYEL/ws
wuHqctRlw/t4extJFYvzASOkBr4zFceR9jCSWR8kOkWY81evx/bxG+eQBJMkzrdw
LOChedVDVIuoTZxfqNzU4Y2TgMGMRWsrEVvBvTMIY3qBue1yeT9M4jzAPhms3/It
Ps7ZeqmF+HrRtFz5ctHQa0QOdZodsKJO0WwjjzjYTDMzZO+bnVFFUy9cG+Gr6mt1
XMKJKkvXQuYTfbRrox4HzIjyfi54xYHnUI35uUUUzEO19Qtm4Ds+sz7/vyz3cYFI
d8IgKoqstjsxtaRq1IS6WIj0bQ/nEqoTNg0I3bndrmCq5LbCoq0z2yXYr5Vl5Gvf
ffbrVM+I91v3R03Svv2Nte2xdbx1RmoI/y3tMyZL
-----END CERTIFICATE-----
`

// NewDefaultAppConfig returns the default app config.
func NewDefaultAppConfig() AppConfig {
	return AppConfig{
		Backend: Backend{
			BitcoinP2PKHActive:       false,
			BitcoinP2WPKHP2SHActive:  true,
			BitcoinP2WPKHActive:      false,
			LitecoinP2WPKHP2SHActive: true,
			LitecoinP2WPKHActive:     false,
			BitcoreP2PKHActive:       false,
                        BitcoreP2WPKHP2SHActive:  true,
                        BitcoreP2WPKHActive:      false,
			EthereumActive:           true,
			BTC: btcCoinConfig{
				ElectrumServers: []*rpc.ServerInfo{
					{
						Server:  "btc.shiftcrypto.ch:443",
						TLS:     true,
						PEMCert: shiftRootCA,
					},
					{
						Server:  "merkle.shiftcrypto.ch:443",
						TLS:     true,
						PEMCert: shiftRootCA,
					},
				},
			},
			TBTC: btcCoinConfig{
				ElectrumServers: []*rpc.ServerInfo{
					{
						Server:  "btc.shiftcrypto.ch:51002",
						TLS:     true,
						PEMCert: shiftRootCA,
					},
					{
						Server:  "merkle.shiftcrypto.ch:51002",
						TLS:     true,
						PEMCert: shiftRootCA,
					},
				},
			},
			LTC: btcCoinConfig{
				ElectrumServers: []*rpc.ServerInfo{
					{
						Server:  "ltc.shiftcrypto.ch:443",
						TLS:     true,
						PEMCert: shiftRootCA,
					},
					{
						Server:  "ltc.shamir.shiftcrypto.ch:443",
						TLS:     true,
						PEMCert: shiftRootCA,
					},
				},
			},
			TLTC: btcCoinConfig{
				ElectrumServers: []*rpc.ServerInfo{
					{
						Server:  "ltc.shiftcrypto.ch:51004",
						TLS:     true,
						PEMCert: shiftRootCA,
					},
					{
						Server:  "ltc.shamir.shiftcrypto.ch:51004",
						TLS:     true,
						PEMCert: shiftRootCA,
					},
				},
			},
			BTX: CoinConfig{
				ElectrumServers: []*rpc.ServerInfo{
					{
						Server:  "btx.shiftcrypto.ch:443",
						TLS:     true,
						PEMCert: shiftRootCA,
					},
					{
						Server:  "btx.shamir.shiftcrypto.ch:443",
						TLS:     true,
						PEMCert: shiftRootCA,
					},
				},
			},
			TBTX: CoinConfig{
				ElectrumServers: []*rpc.ServerInfo{
					{
						Server:  "btx.shiftcrypto.ch:51004",//TODO BTX
						TLS:     true,
						PEMCert: shiftRootCA,
					},
					{
						Server:  "btx.shamir.shiftcrypto.ch:51004",//TODO BTX
						TLS:     true,
						PEMCert: shiftRootCA,
					},
				},
			},
			ETH: ethCoinConfig{
				NodeURL: "https://mainnet.infura.io/v3/2ce516f67c0b48e8af5387b714ab8a61",
			},
			TETH: ethCoinConfig{
				NodeURL: "https://ropsten.infura.io/v3/2ce516f67c0b48e8af5387b714ab8a61",
			},
			RETH: ethCoinConfig{
				NodeURL: "https://rinkeby.infura.io/v3/2ce516f67c0b48e8af5387b714ab8a61",
			},
		},
	}
}

// Config manages the app configuration.
type Config struct {
	lock locker.Locker

	appConfigFilename string
	appConfig         AppConfig

	accountsConfigFilename string
	accountsConfig         AccountsConfig
}

// NewConfig creates a new Config, stored in the given location. The filename must be writable, but
// does not have to exist.
func NewConfig(appConfigFilename string, accountsConfigFilename string) *Config {
	config := &Config{
		appConfigFilename: appConfigFilename,
		appConfig:         NewDefaultAppConfig(),

		accountsConfigFilename: accountsConfigFilename,
		accountsConfig:         newDefaultAccountsonfig(),
	}
	config.load()
	return config
}

func (config *Config) load() {
	jsonBytes, err := ioutil.ReadFile(config.appConfigFilename)
	if err != nil {
		return
	}
	if err := json.Unmarshal(jsonBytes, &config.appConfig); err != nil {
		return
	}
	jsonBytes, err = ioutil.ReadFile(config.accountsConfigFilename)
	if err != nil {
		return
	}
	if err := json.Unmarshal(jsonBytes, &config.accountsConfig); err != nil {
		return
	}
}

// AppConfig returns the app config.
func (config *Config) AppConfig() AppConfig {
	defer config.lock.RLock()()
	return config.appConfig
}

// SetAppConfig sets and persists the app config.
func (config *Config) SetAppConfig(appConfig AppConfig) error {
	defer config.lock.Lock()()
	config.appConfig = appConfig
	return config.save(config.appConfigFilename, config.appConfig)
}

// AccountsConfig returns the accounts config.
func (config *Config) AccountsConfig() AccountsConfig {
	defer config.lock.RLock()()
	return config.accountsConfig
}

// SetAccountsConfig sets and persists the accounts config.
func (config *Config) SetAccountsConfig(accountsConfig AccountsConfig) error {
	defer config.lock.Lock()()
	config.accountsConfig = accountsConfig
	return config.save(config.accountsConfigFilename, config.accountsConfig)
}

func (config *Config) save(filename string, conf interface{}) error {
	jsonBytes, err := json.MarshalIndent(conf, "", "    ")
	if err != nil {
		return errp.WithStack(err)
	}
	return errp.WithStack(ioutil.WriteFile(filename, jsonBytes, 0644))
}
