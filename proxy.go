package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type Proxy struct {
	target_url     string
	listening_port string
}

func main() {

	proxy := Proxy{
		target_url:     "http://stackoverflow.com",
		listening_port: "8080",
	}
	err := StartProxy(&proxy)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func StartProxy(p *Proxy) error {

	http.HandleFunc("/", p.ProxyRequest)
	log.Println("Started GO proxyserver on port", p.listening_port)

	err := http.ListenAndServe("127.0.0.1:"+p.listening_port, nil)
	if err != nil {
		return err
	}
	return nil
}

func (p *Proxy) ProxyRequest(w http.ResponseWriter, r *http.Request) {

	uri := p.target_url + r.RequestURI
	log.Println(r.Method + ": " + uri)

	p.MethodHandler(r)

	remote_request, err := p.CreateRemoteRequest(r, uri)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	p.CopyHeader(r.Header, &remote_request.Header)

	resp, err := p.Query(remote_request)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := p.ReadBody(resp)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	destination_header := w.Header()
	p.CopyHeader(resp.Header, &destination_header)
	destination_header.Add("Requested-Host", remote_request.Host)
	w.Write(body)
}

func (p *Proxy) MethodHandler(r *http.Request) {
	if r.Method == "POST" {
		log.Printf("Method is POST:")
	}
}

func (p *Proxy) CreateRemoteRequest(r *http.Request, uri string) (*http.Request, error) {
	rr, err := http.NewRequest(r.Method, uri, r.Body)
	if err != nil {
		return nil, err
	}
	return rr, nil
}

func (p *Proxy) Query(r *http.Request) (*http.Response, error) {
	// Create a client and query the target
	var transport http.Transport
	resp, err := transport.RoundTrip(r)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (p *Proxy) ReadBody(r *http.Response) ([]byte, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (p *Proxy) CopyHeader(source http.Header, dest *http.Header) {
	for n, v := range source {
		for _, vv := range v {
			dest.Add(n, vv)
		}
	}
}
