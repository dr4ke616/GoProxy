package server

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

// Struct to defin the config file. Represented using JSON
type Proxy struct {
	ListeningPort  string `json:"listening_port"`
	TargetUrl      string `json:"target_url"`
	RoutingOptions []struct {
		URI           string         `json:"uri"`
		FromMethod    string         `json:"from_method"`
		ToMethod      string         `json:"to_method"`
		CustomHeaders []CustomHeader `json:"custom_headers"`
	} `json:"routing_options"`
}

type CustomHeader struct {
	Replace      bool     `json:"replace"`
	HeaderKey    string   `json:"header_key"`
	HeaderValues []string `json:"header_values"`
}

// Struct defining the rules for the route handeling
type RouteHandler struct {
	FromMethod, ToMethod string
	CustomHeaders        []CustomHeader
	Proxy                *Proxy
}

// HTTP interface to be overloaded
type Handler interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}

// Start a proxy webserver, listening on the port specified in the
// config. All traffic will be routed to the target URL. Any custom
// headers or metod types will be handled
func StartProxy(p *Proxy) error {

	// Handle the custom routing options
	for _, route := range p.RoutingOptions {
		log.Println("Adding custom handler for URI", route.URI)
		handler := RouteHandler{
			FromMethod:    route.FromMethod,
			ToMethod:      route.ToMethod,
			CustomHeaders: route.CustomHeaders,
			Proxy:         p,
		}
		http.Handle(route.URI, Handler(handler))
	}

	// Handle the default root url handler
	http.Handle("/", Handler(RouteHandler{
		FromMethod:    "",
		ToMethod:      "",
		CustomHeaders: nil,
		Proxy:         p,
	}))

	// Lets Go...
	log.Println("Starting GO proxyserver on port", p.ListeningPort)
	err := http.ListenAndServe("127.0.0.1:"+p.ListeningPort, nil)
	if err != nil {
		return err
	}
	return nil
}

// Handele the incomeing requests and re-route to the target
func (h RouteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.Method == h.FromMethod {
		r.Method = h.ToMethod
	}

	uri := h.Proxy.TargetUrl + r.RequestURI
	log.Println(r.Method + ": " + uri)

	remote_request, err := CreateRemoteRequest(r, uri)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	CopyHeader(r.Header, &remote_request.Header)

	resp, err := Query(remote_request)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := ReadBody(resp)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// Build the headers to be sent to the client
	destination_header := w.Header()
	CopyHeader(resp.Header, &destination_header)
	destination_header.Add("Requested-Host", remote_request.Host)
	h.HandleCustomHeaders(&destination_header)
	w.Write(body)
}

// Handles any custom headers that are specified in the config
func (h RouteHandler) HandleCustomHeaders(destination_header *http.Header) {

	for _, header := range h.CustomHeaders {
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

// Create a client and query the target
func Query(r *http.Request) (*http.Response, error) {

	var transport http.Transport
	resp, err := transport.RoundTrip(r)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Reads the body from the target endpoint
func ReadBody(r *http.Response) ([]byte, error) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func CopyHeader(source http.Header, dest *http.Header) {

	for n, v := range source {
		for _, vv := range v {
			dest.Add(n, vv)
		}
	}
}
