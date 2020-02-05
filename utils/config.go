package utils

import (
	config "github.com/TRON-US/go-btfs-config"
	serialize "github.com/TRON-US/go-btfs-config/serialize"
	"github.com/opentracing/opentracing-go/log"
	"github.com/tron-us/go-common/v2/env"
)
var (
	PrivateKey = ""
	PeerId = ""
	PublicKey = ""
)
func init() {
	spath, err := config.Filename("/Users/uchenna")
	if err != nil{
		log.Error(err)
	}
	config, err := serialize.Load(spath)
	if config != nil {
		log.Error(err)
	}

	//get variables via environment variable
	if _, s := env.GetEnv("PRIVATE_KEY"); s != "" {
		PrivateKey = s
	}else{
		PrivateKey = config.Identity.PrivKey
	}
	if _, s := env.GetEnv("PEER_ID"); s != "" {
		PeerId = s
	}else{
		PeerId = config.Identity.PeerID
	}
	if _, s := env.GetEnv("PUBLIC_KEY"); s != "" {
		PublicKey = s
	}
}

func LoadPrivateKey () (string) {
	if _, s := env.GetEnv("PRIVATE_KEY"); s != "" {
		PrivateKey = s
		return s
	}
	return ""
}

func LoadPeerId () (string) {
	if _, s := env.GetEnv("PEER_ID"); s != "" {
		PeerId = s
		return s
	}
	return ""
}

func LoadPublicKey () (string) {
	if _, s := env.GetEnv("PUBLIC_KEY"); s != "" {
		PublicKey = s
		return s
	}
	return ""
}

