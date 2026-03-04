package main

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type HostAuth struct {
	Endpoint, Key, User, Pass string
}

type HostConfig struct {
	Auth HostAuth
}

type HostCredentials struct {
	User, Pass string
}

type Config struct {
	Hosts map[string]HostConfig
}

// search yaml file for an entry exactly matching `host`
// `host`.login_path
// nil if endpoint is not found
func getAuthenticationEndpoint(host string) (string, error) {
	stdPath := "/api/auth/login"
	rootFs := os.DirFS(".")
	log.Println("Reading config.yaml...")
	ymlBytes, err := fs.ReadFile(rootFs, "config.yaml")
	if err != nil {
		log.Println("Cannot read config.yml", err)
	}

	if len(ymlBytes) == 0 {
		log.Println("Config file is empty, returning the standard login path")
		return stdPath, nil
	}

	var yamlData Config
	yaml.Unmarshal(ymlBytes, &yamlData)
	if err != nil {
		log.Fatal("Cannot unmarshal config.yaml", err)
	}

	uriConfig, ok := yamlData.Hosts[host]
	if !ok {
		return "/api/auth/login2", nil
	}

	return uriConfig.Auth.Endpoint, nil
}

func getAuthenticationCredentials(host string) HostCredentials {
	return HostCredentials{"sistemas@devppay.com.br", "AA11QQ11"}
}

func handleRoot(rw http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	log.Println("Proxying the request...")

	// get destination URL
	// host := req.Host
	apiPath, err := getAuthenticationEndpoint(req.Host)
	if err != nil {
		log.Println(err)
	}

	authCredentials := getAuthenticationCredentials(req.Host)
	authEndpoint := fmt.Sprintf("http://%s%s", req.Host, apiPath)
	authPayload := fmt.Sprintf("{\"email\":\"%s\", \"password\":\"%s\"}", authCredentials.User, authCredentials.Pass)
	log.Printf("Payload is: %s", authPayload)

	log.Printf("trying to log in at host auth endpoint")
	authResponse, err := http.Post(authEndpoint, "application/json", strings.NewReader(authPayload))
	if err != nil {
		log.Fatal("Error loging in ", err)
	}

	authResponseBody, err := io.ReadAll(authResponse.Body)
	defer authResponse.Body.Close()
	if err != nil {
		log.Fatal("Cannot read login body", err)
	}

	log.Printf("%s\n", authResponseBody)

	// get bearer token from destination login endpoint
	// save token in cache (file((fallback when memcached is down))/memcached) [log result]
	bearerToken := fmt.Sprintf("Bearer %s", authResponseBody)

	proxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetXForwarded()
			r.Out.Header.Add("Accept", "application/json")
			r.Out.Header.Add("Content-Type", "application/json")
			r.Out.Header.Add("Authentication", bearerToken)
		},
	}

	proxy.ServeHTTP(rw, req)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("No environment file configured!!")
	}

	log.Println("Loaded environment file.")

	http.HandleFunc("/", handleRoot)

	log.Fatal(http.ListenAndServe(":1080", nil))
}
