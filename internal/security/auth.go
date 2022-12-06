package security

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"reflect"
	"strings"

	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"

	"net/http"
)

type Claims map[string]interface{}
type openArray []interface{}

const wellKnown = ".well-known/openid-configuration"

var (
	wellKnownCache = make(map[string]map[string]interface{})
	jwksCache      = make(map[string]*jose.JSONWebKeySet)
)

func fetchWellKnown(iss string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", iss, wellKnown), nil)
	if err != nil {
		return nil, fmt.Errorf("could not create well-known request: %w", err)
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not fetch well-known: %w", err)
	}
	defer closeBody(res)

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("received non-200 response code")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response body: %w", err)
	}

	urls := make(map[string]interface{})
	err = json.Unmarshal(body, &urls)
	if err != nil {
		return nil, fmt.Errorf("could not parse response body: %w", err)
	}

	return urls, nil
}

func fetchJwks(iss string) (*jose.JSONWebKeySet, error) {

	var err error
	urls, ok := wellKnownCache[iss]
	if !ok || urls == nil {
		urls, err = fetchWellKnown(iss)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest("GET", urls["jwks_uri"].(string), nil)
	if err != nil {
		return nil, fmt.Errorf("could not create jwks request: %w", err)
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not fetch jwks: %w", err)
	}
	defer closeBody(res)

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("received non-200 response code")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response body: %w", err)
	}

	jwks := jose.JSONWebKeySet{}
	err = json.Unmarshal(body, &jwks)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal jwks into struct: %w", err)
	}

	return &jwks, nil
}

func verifyToken(bearerToken string) (map[string]interface{}, error) {
	token, err := jwt.ParseSigned(bearerToken)
	if err != nil {
		return nil, fmt.Errorf("could not parse Bearer token: %w", err)
	}

	out := make(map[string]interface{})
	err = token.UnsafeClaimsWithoutVerification(&out)
	if err != nil {
		return nil, fmt.Errorf("unable to extract claims: %w", err)
	}

	iss, ok := out["iss"]
	if !ok {
		return nil, fmt.Errorf("missing iss claim")
	}
	issuer, ok := iss.(string)
	if !ok {
		return nil, fmt.Errorf("iss claim is invalid")
	}

	// Get jwks
	jsonWebKeySet, ok := jwksCache[issuer]
	if !ok {
		jsonWebKeySet, err = fetchJwks(issuer)
		if err != nil {
			return nil, fmt.Errorf("could not load JWKS: %w", err)
		}
	}

	// Get claims out of token (validate signature while doing this)
	claims := make(map[string]interface{})
	err = token.Claims(jsonWebKeySet, &claims)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve claims: %w", err)
	}

	return claims, nil
}

func extractCredentials(credentials string) (map[string]interface{}, error) {
	bs, err := base64.URLEncoding.DecodeString(credentials)
	if err != nil {
		return nil, fmt.Errorf("unable to decode credentials: %w", err)
	}

	parts := strings.SplitN(string(bs), ":", 2)
	claims := make(map[string]interface{})
	claims["username"] = parts[0]
	if len(parts) > 1 {
		claims["password"] = parts[1]
		claims[parts[0]] = parts[1]
	} else {
		claims["password"] = ""
		claims[parts[0]] = ""
	}

	return claims, nil
}

func BearerAuthorized(r *http.Request, checks Claims) (bool, Claims) {
	auth := r.Header.Get("Authorization")
	if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
		return false, Claims{}
	}
	claims, err := verifyToken(auth[7:])
	if err != nil {
		return false, Claims{}
	}

	return validate(checks, claims), claims
}

func CredentialsAuthorized(r *http.Request, checks Claims) (bool, Claims) {
	auth := r.Header.Get("Authorization")
	if auth == "" || !strings.HasPrefix(auth, "Basic ") {
		return false, Claims{}
	}

	claims, err := extractCredentials(auth[6:])
	if err != nil {
		return false, Claims{}
	}

	return validate(checks, claims), claims
}

func validate(checks Claims, claims Claims) bool {
	for key, expected := range checks {
		actual, ok := claims[key]
		if !ok {
			return false
		}
		te := reflect.TypeOf(expected)
		ta := reflect.TypeOf(actual)
		if te.Kind() != ta.Kind() {
			return false
		}
		switch ev := expected.(type) {
		case string:
			if expected.(string) != actual.(string) {
				return false
			}
		case []interface{}:
			av := openArray(actual.([]interface{}))
			if len(ev) == 0 {
				return false
			}
			for _, v := range ev {
				if !av.contains(v) {
					return false
				}
			}
		default:
			fmt.Printf("Other: %v", ev)
			return false
		}
	}
	return true
}

func (o openArray) contains(v interface{}) bool {
	for _, a := range o {
		if a == v {
			return true
		}
	}
	return false
}

func closeBody(res *http.Response) {
	if res != nil && res.Body != nil {
		err := res.Body.Close()
		if err != nil {
			log.Printf("failed to close response body: %v", err)
		}
	}
}
