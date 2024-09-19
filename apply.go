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

	"github.com/rs/zerolog/log"
)

func fromCmdline(s string) map[string]string {
	log.Debug().Str("arg", s).Msg("parse -param.kvs argument")

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
		log.Trace().Str("key", k).Str("value", v).Msg("storing command line kv")
		kvMap[k] = v
		k = ""
	}

	return kvMap
}

func crossReference(template Template, kvs map[string]string, groupId string) map[string]string {
	log.Debug().Msg("cross referencing kvs")

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
	log.Trace().Str("s", s).Msg("find keys")
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
		key := getPrivateKey(*params.privateKeyPath)
		sig := signPayload(key, apply)
		apply = apply + "&sig=" + sig
	}

	fmt.Printf("%s/v2/domainTemplates/providers/%s/services/%s/apply?", settings.URLSyncUX, template.ProviderID, template.ServiceID)
	fmt.Printf("%s\n", apply)
}

func getPrivateKey(pathToKey string) any {
	keyBytes, err := os.ReadFile(pathToKey)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot read file")
	}
	keyBlock, _ := pem.Decode(keyBytes)
	if keyBlock == nil {
		log.Fatal().Err(err).Str("file", pathToKey).Msg("cannot decode key")
	}
	key, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if err != nil {
		log.Fatal().Err(err).Str("file", pathToKey).Msg("cannot parse x509")
	}
	return key
}

func signPayload(key any, apply string) string {
	msgHash := sha256.New()
	_, err := msgHash.Write([]byte(apply))
	if err != nil {
		log.Fatal().Err(err).Msg("could not generate hash")
	}
	hashSum := msgHash.Sum(nil)
	var signature []byte
	if privateKey, ok := key.(*rsa.PrivateKey); ok {
		signature, err = rsa.SignPKCS1v15(crand.Reader, privateKey, crypto.SHA256, hashSum)
		if err != nil {
			log.Fatal().Err(err).Msg("rsa sign failed")
		}
	} else {
		log.Fatal().Msg("not a rsa key")
	}
	return base64.StdEncoding.EncodeToString(signature)
}
