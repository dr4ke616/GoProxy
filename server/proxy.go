package server

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

var DEVIL = false
var ALLOWED_METHODS = [4]string{"GET", "POST", "PUT", "PATCH"}

// Struct to defin the config file. Represented using JSON
type Proxy struct {
	LogFile        string `json:"log_file"`
	ListeningPort  string `json:"listening_port"`
	TargetUrl      string `json:"target_url"`
	RoutingOptions []struct {
		URI           string         `json:"uri"`
		FromMethod    string         `json:"from_method"`
		ToMethod      string         `json:"to_method"`
		CustomHeaders []CustomHeader `json:"custom_headers"`
	} `json:"routing_options"`
	Transport http.Transport
}

type CustomHeader struct {
	Replace      bool     `json:"replace"`
	HeaderKey    string   `json:"header_key"`
	HeaderValues []string `json:"header_values"`
}

// Struct defining the rules for the route handeling
type CustomHandler struct {
	FromMethod, ToMethod string
	CustomHeaders        []CustomHeader
	Active               bool
}

// Start a proxy webserver, listening on the port specified in the
// config. All traffic will be routed to the target URL. Any custom
// headers or metod types will be handled
func StartProxy(p *Proxy) error {

	if !DEVIL {
		p.HandleLogging()
	}

	http.HandleFunc("/", p.ServeHTTP)

	// Lets Go...
	log.Println("Starting GO proxyserver on port", p.ListeningPort)
	err := http.ListenAndServe("127.0.0.1:"+p.ListeningPort, nil)
	if err != nil {
		return err
	}
	return nil
}

// Handle the incomeing requests and re-route to the target
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	full_url := p.TargetUrl + r.RequestURI

	custom_handler := CustomHandler{Active: false}
	p.InitCustomHandler(r, &custom_handler)
	if custom_handler.Active {
		log.Println("Handeling custom route for", full_url)
	}

	if err := HandleCustomMethod(r, &custom_handler); err != nil {
		panic(err)
	}

	remote_request, err := CreateRemoteRequest(r, full_url)
	if err != nil {
		panic(err)
	}
	CopyHeader(r.Header, &remote_request.Header)

	resp, err := p.Query(remote_request)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ReadBody(resp)
	if err != nil {
		panic(err)
	}

	// Build the headers to be sent to the client
	destination_header := w.Header()
	CopyHeader(resp.Header, &destination_header)
	destination_header.Add("Requested-Host", remote_request.Host)
	HandleCustomHeaders(&destination_header, &custom_handler)

	w.WriteHeader(resp.StatusCode)
	w.Write(body)

	log.Println(r.Method + ": " + full_url)
}

func (p *Proxy) InitCustomHandler(r *http.Request, c *CustomHandler) {

	for _, route := range p.RoutingOptions {
		uri := r.RequestURI

		route.URI = strings.Split(route.URI, "?")[0]
		s := strings.Split(uri, "?")
		if len(s) > 1 {
			// params := s[1]
			uri = s[0]
		}

		if uri == route.URI {
			c.FromMethod = route.FromMethod
			c.ToMethod = route.ToMethod
			c.CustomHeaders = route.CustomHeaders
			c.Active = true
			return
		}
	}
}

// Create a client and query the target
func (p *Proxy) Query(r *http.Request) (*http.Response, error) {

	resp, err := p.Transport.RoundTrip(r)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Switches the method type as specified in the config
func HandleCustomMethod(r *http.Request, c *CustomHandler) error {

	if !c.Active {
		return nil
	}

	var err error

	if err = ValidateMethod(c.FromMethod); err != nil {
		return err
	}

	if err = ValidateMethod(c.ToMethod); err != nil {
		return err
	}

	if c.FromMethod == "" || c.ToMethod == "" {
		return nil
	}

	if r.Method == c.FromMethod {
		r.Method = c.ToMethod
	}
	return nil
}

// Handles any custom headers that are specified in the config
func HandleCustomHeaders(destination_header *http.Header, c *CustomHandler) {

	if !c.Active {
		return
	}

	for _, header := range c.CustomHeaders {
		if header.Replace {
			// When we replace we remove the old header and add the new one
			destination_header.Set(header.HeaderKey, strings.Join(header.HeaderValues, ", "))
		} else {
			// Otherwise we just append onto the already existing header
			new_header := destination_header.Get(header.HeaderKey) +
				", " + strings.Join(header.HeaderValues, ", ")
			destination_header.Set(header.HeaderKey, new_header)
		}
	}
}

// Creates the remote request object
func CreateRemoteRequest(r *http.Request, uri string) (*http.Request, error) {

	rr, err := http.NewRequest(r.Method, uri, r.Body)
	if err != nil {
		return nil, err
	}
	return rr, nil
}

// Reads the body from the target endpoint
func ReadBody(r *http.Response) ([]byte, error) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// Used to copy headers from the target to the client
func CopyHeader(source http.Header, dest *http.Header) {

	for n, v := range source {
		for _, vv := range v {
			dest.Add(n, vv)
		}
	}
}

// Verify the methods are correct
func ValidateMethod(method string) error {

	for _, m := range ALLOWED_METHODS {
		if method == m {
			return nil
		}
	}

	return fmt.Errorf("Method type %s is not allowed", method)
}

func (p *Proxy) HandleLogging() error {
	f, err := os.OpenFile(p.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	log.SetOutput(f)
	return nil
}
