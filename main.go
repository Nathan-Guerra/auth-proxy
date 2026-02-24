package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type HostAuth struct {
	Endpoint, Key, User, Pass string
}

type HostConfig struct {
	Auth HostAuth
}

type Config struct {
	Hosts map[string]HostConfig
}

// search yaml file for an entry exactly matching `host`
// `host`.login_path
// nil if endpoint is not found
func getAuthenticationEndpoint(host string) (string, error) {
	stdPath := "/api/auth/login"

	log.Println("Opening current directory...")
	rootFs := os.DirFS(".")
	fmt.Printf("[debug] Root dir is:  %+v\n", rootFs)

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

	jsonYaml, err := json.Marshal(yamlData)
	if err != nil {
		log.Fatal("unable to marshal yaml config.", err)
	}

	fmt.Printf("json yaml: %+v\n", string(jsonYaml))

	uriConfig, ok := yamlData.Hosts[host]

	if !ok {
		return "/api/auth/login2", nil
	}

	return uriConfig.Auth.Endpoint, nil

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

	log.Println("auth endpoint is ", apiPath)

	// get bearer token from destination login endpoint
	// save token in cache (file((fallback when memcached is down))/memcached) [log result]

	proxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetXForwarded()
			r.Out.Header.Add("Authentication", "Bearer "+"131")
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
