package server

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

var DEVIL = false

// Struct to defin the config file. Represented using JSON
type Proxy struct {
	LogFile       string `json:"log_file"`
	ListeningPort string `json:"listening_port"`
	TargetUrl     string `json:"target_url"`
	SSL           struct {
		Active        bool   `json:"active"`
		KeyFile       string `json:"key_file"`
		CertFile      string `json:"cert_file"`
		ListeningPort string `json:"listening_port"`
	}
	RoutingOptions []struct {
		URI            string         `json:"uri"`
		FromMethod     string         `json:"from_method"`
		ToMethod       string         `json:"to_method"`
		CopyParamaters bool           `json:"copy_paramaters"`
		CustomHeaders  []CustomHeader `json:"custom_headers"`
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
	CopyParamaters       bool
	Paramaters           map[string][]string
	RawData              string
}

// Start a proxy webserver, listening on the port specified in the
// config. All traffic will be routed to the target URL. Any custom
// headers or metod types will be handled
func StartProxy(p *Proxy) error {

	if !DEVIL {
		p.handleLogging()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", p.ServeHTTP)

	// If SSL is activated we can listen on a port for HTTPS connections
	if p.SSL.Active {
		log.Println("Starting SSL GO proxyserver on port", p.SSL.ListeningPort)

		go func() {
			err := http.ListenAndServeTLS("127.0.0.1:"+p.SSL.ListeningPort, p.SSL.CertFile, p.SSL.KeyFile, mux)
			if err != nil {
				panic(err)
			}
		}()
	}

	// Lets Go...
	log.Println("Starting GO proxyserver on port", p.ListeningPort)
	err := http.ListenAndServe("127.0.0.1:"+p.ListeningPort, mux)
	if err != nil {
		return err
	}
	return nil
}

// Handle the incomeing requests and re-route to the target
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	logIncomingRequest(r)

	custom_handler := CustomHandler{Active: false}
	if err := p.initCustomHandler(r, &custom_handler); err != nil {
		panic(err)
	}

	if err := p.copyParamaters(r, &custom_handler); err != nil {
		panic(err)
	}

	if err := p.handleCustomMethod(r, &custom_handler); err != nil {
		panic(err)
	}

	if err := p.logActivity(r, &custom_handler); err != nil {
		panic(err)
	}

	full_url, err := p.fullURL(r)
	if err != nil {
		panic(err)
	}

	remote_request, err := createRemoteRequest(r, full_url.String())
	if err != nil {
		panic(err)
	}

	copyHeader(r.Header, &remote_request.Header)

	resp, err := p.query(remote_request)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := readBody(resp)
	if err != nil {
		panic(err)
	}

	// Build the headers to be sent to the client
	destination_header := w.Header()
	copyHeader(resp.Header, &destination_header)
	destination_header.Add("Requested-Host", remote_request.Host)
	handleCustomHeaders(&destination_header, &custom_handler)

	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

func (p *Proxy) fullURL(r *http.Request) (*url.URL, error) {
	return url.Parse(p.TargetUrl + r.RequestURI)
}

func (p *Proxy) initCustomHandler(r *http.Request, c *CustomHandler) error {

	targetEndpoint, err := p.fullURL(r)
	params, err := url.ParseQuery(targetEndpoint.RawQuery)
	if err != nil {
		return err
	}

	for _, customRoute := range p.RoutingOptions {
		customRoute.URI = strings.Split(customRoute.URI, "?")[0]

		if targetEndpoint.Path == customRoute.URI {
			c.FromMethod = customRoute.FromMethod
			c.ToMethod = customRoute.ToMethod
			c.CopyParamaters = customRoute.CopyParamaters
			c.Paramaters = params
			c.CustomHeaders = customRoute.CustomHeaders
			c.Active = true
			return nil
		}
	}
	return nil
}

// Create a client and query the target
func (p *Proxy) query(r *http.Request) (*http.Response, error) {

	resp, err := p.Transport.RoundTrip(r)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Switches the method type as specified in the config
func (p *Proxy) handleCustomMethod(r *http.Request, c *CustomHandler) error {

	if c.Active {
		r.Method = c.ToMethod
	}
	return nil
}

func (p *Proxy) copyParamaters(r *http.Request, c *CustomHandler) error {

	if !c.Active || !c.CopyParamaters {
		return nil
	}

	targetEndpoint, err := p.fullURL(r)
	if err != nil {
		return err
	}

	r.RequestURI = targetEndpoint.Path
	for _, customHeader := range c.CustomHeaders {
		for _, h := range customHeader.HeaderValues {
			h := strings.ToLower(h)
			if h == "application/json" {
				if err := handleApplicationJson(r, c); err != nil {
					return err
				}
				return nil
			}
			if h == "application/xml" {
				if err := handleApplicationXML(r, c); err != nil {
					return err
				}
				return nil
			}
			if h == "application/x-www-form-urlencoded" {
				if err := handleApplicationForm(r, c); err != nil {
					return err
				}
				return nil
			}
		}
	}
	return nil
}

func handleApplicationJson(r *http.Request, c *CustomHandler) error {

	params := make(map[string]interface{}, len(c.Paramaters))
	for k, v := range c.Paramaters {

		// First we check if its an int
		// Then check for boolean, finally
		// if none of them its a string
		values := make([]interface{}, len(v))
		for i, val := range v {
			if _, err := strconv.Atoi(val); err == nil {
				intval, _ := strconv.ParseInt(val, 0, 64)
				values[i] = intval
			} else if val == "true" {
				values[i] = true
			} else if val == "false" {
				values[i] = false
			} else {
				values[i] = val
			}
		}

		// If there is only one occuerance of a
		// value, we use that one. Other wise
		// we set the value for a given key
		// as a list
		if len(values) == 1 {
			params[k] = values[0]
		} else {
			params[k] = values
		}
	}

	jsonString, err := json.Marshal(params)
	if err != nil {
		return err
	}

	r.Body = ioutil.NopCloser(bytes.NewReader(jsonString))
	c.RawData = string(jsonString)
	return nil
}

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

func handleApplicationXML(r *http.Request, c *CustomHandler) error {
	log.Println("Warnng! Not Implemented: copying paramaters for application/xml not implemented yet.")
	return nil
}

func handleApplicationForm(r *http.Request, c *CustomHandler) error {
	params := url.Values{}
	for k, v := range c.Paramaters {
		for _, val := range v {
			params.Add(k, val)
		}
	}
	r.Body = nopCloser{bytes.NewBufferString(params.Encode())}
	c.RawData = params.Encode()
	return nil
}

// Handles any custom headers that are specified in the config
func handleCustomHeaders(destination_header *http.Header, c *CustomHandler) {

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
func createRemoteRequest(r *http.Request, uri string) (*http.Request, error) {

	rr, err := http.NewRequest(r.Method, uri, r.Body)
	if err != nil {
		return nil, err
	}
	return rr, nil
}

// Reads the body from the target endpoint
func readBody(r *http.Response) ([]byte, error) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// Used to copy headers from the target to the client
func copyHeader(source http.Header, dest *http.Header) {

	for n, v := range source {
		for _, vv := range v {
			dest.Add(n, vv)
		}
	}
}

func (p *Proxy) handleLogging() error {
	f, err := os.OpenFile(p.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	log.SetOutput(f)
	return nil
}

func (p *Proxy) logActivity(r *http.Request, c *CustomHandler) error {

	full_url, err := p.fullURL(r)
	if err != nil {
		return err
	}

	if c.Active {
		msg := "Handeling custom route for " + full_url.String()
		if c.CopyParamaters {
			msg += " with data " + c.RawData
		}
		log.Println(msg)
	}
	log.Println("Proxy request to: [" + r.Method + "] " + full_url.String())
	return nil
}

func logIncomingRequest(r *http.Request) {

	var scheme string
	if r.TLS != nil {
		scheme = "https"
	} else {
		scheme = "http"
	}
	log.Println("Incoming request: [" + r.Method + "] " + scheme + "://" + r.Host + r.URL.String())
}
