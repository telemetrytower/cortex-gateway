package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/golang-jwt/jwt/v4"
	"github.com/rewe-digital/cortex-gateway/pkg/org"
)

var (
	tenantID  string
	aud       string
	version   uint
	jwtSecret string
)

func main() {
	flag.StringVar(&tenantID, "tenant.id", "", "The tenant of JSON Web Token")
	flag.StringVar(&aud, "tenant.aud", "", "The audience of JSON Web Token")
	flag.UintVar(&version, "tenant.version", 1, "The version of JSON Web Token")
	flag.StringVar(&jwtSecret, "auth.jwt-secret", "", "Secret to sign JSON Web Token")
	flag.Parse()

	if jwtSecret == "" {
		log.Panic("empty jwt secret")
	}

	tenant := org.Tenant{
		TenantID: tenantID,
		Audience: aud,
		Version:  uint8(version),
	}

	if err := tenant.Valid(); err != nil {
		log.Panicf("tenant valid(): %v", err)
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS512, &tenant).SignedString([]byte(jwtSecret))
	if err != nil {
		log.Panicf("jwt.SigningMethodHS512.Sign(): %v", err)
	}

	fmt.Println(token)
}
