package main

import (
	"crypto"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/rand/v2"
	"os"
	"strings"
)

func fromCmdline(s string) map[string]string {
	kvMap := make(map[string]string)
	parts := strings.Split(s, "%")

	var k string
	for _, v := range parts {
		if len(v) < 1 {
			continue
		}
		if k == "" {
			k = v
			continue
		}
		kvMap[k] = v
		k = ""
	}

	return kvMap
}

func crossReference(template Template, kvs map[string]string, groupId string) map[string]string {
	checked := make(map[string]string)

	findKeys(template.ProviderName, kvs, &checked)
	findKeys(template.ServiceName, kvs, &checked)
	findKeys(template.Logo, kvs, &checked)

	for _, record := range template.Records {
		if groupId != "" && groupId != record.GroupID {
			continue
		}
		findKeys(record.Type, kvs, &checked)
		findKeys(record.Host, kvs, &checked)
		findKeys(record.Name, kvs, &checked)
		findKeys(record.PointsTo, kvs, &checked)
		findKeys(record.Data, kvs, &checked)
		findKeys(record.Service, kvs, &checked)
		findKeys(record.Target, kvs, &checked)
		findKeys(record.SPFRules, kvs, &checked)
	}

	return checked
}

func findKeys(s string, kvs map[string]string, checked *map[string]string) {
	parts := strings.Split(s, "%")

	for i, key := range parts {
		if i%2 == 0 {
			continue
		}
		if val, ok := kvs[key]; ok {
			(*checked)[key] = val
		} else {
			(*checked)[key] = randSeq(8)
		}
	}
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.N(len(letters))]
	}
	return string(b)
}

func printApply(fqdn string, template Template, settings Settings, params *parameters, kvs map[string]string) {
	fmt.Printf("%s/v2/domainTemplates/providers/%s/services/%s/apply?", settings.URLSyncUX, template.ProviderID, template.ServiceID)

	apply := "domain=" + fqdn

	if *params.host != "" {
		apply = apply + "&host=" + *params.host
	}
	if *params.redirect_url != "" {
		apply = apply + "&redirect_uri=" + *params.redirect_url
	}
	if *params.state != "" {
		apply = apply + "&state=" + *params.state
	} else {
		apply = apply + "&state=" + randSeq(8)
	}
	if *params.providerName != "" {
		apply = apply + "&providerName=" + *params.providerName
	}
	if *params.serviceName != "" {
		apply = apply + "&serviceName=" + *params.serviceName
	}
	if *params.groupId != "" {
		apply = apply + "&groupId=" + *params.groupId
	}
	if *params.key != "" {
		apply = apply + "&key=" + *params.key
	}

	// signature generation must happen last
	if params.privateKeyPath != nil {
		key, err := getPrivateKey(*params.privateKeyPath)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		sig, err := signPayload(key, apply)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		apply = apply + "&sig=" + sig
	}

	fmt.Printf("%s\n", apply)
}

func getPrivateKey(pathToKey string) (any, error) {
	keyBytes, err := os.ReadFile(pathToKey)
	if err != nil {
		return nil, err
	}
	keyBlock, _ := pem.Decode(keyBytes)
	if keyBlock == nil {
		return nil, err
	}
	key, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func signPayload(key any, apply string) (string, error) {
	msgHash := sha256.New()
	_, err := msgHash.Write([]byte(apply))
	if err != nil {
		return "", err
	}
	hashSum := msgHash.Sum(nil)
	var signature []byte
	if privateKey, ok := key.(*rsa.PrivateKey); ok {
		signature, err = rsa.SignPKCS1v15(crand.Reader, privateKey, crypto.SHA256, hashSum)
		if err != nil {
			return "", err
		}
	} else {
		return "", fmt.Errorf("not rsa key")
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}
