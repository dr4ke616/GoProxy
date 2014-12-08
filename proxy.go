package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type Target struct {
	URL string
}

func main() {
	log.Fatal(StartServer("http://stackoverflow.com"))
}

func StartServer(target_url string, port ...string) error {
	var listening_port = "8080"
	if len(port) > 0 {
		listening_port = port[0]
	}

	target := Target{URL: target_url}
	http.HandleFunc("/", target.ProxyRequest)
	log.Println("Started GO proxyserver on port", listening_port)

	err := http.ListenAndServe("127.0.0.1:"+listening_port, nil)
	if err != nil {
		return err
	}
	return nil
}

func (t *Target) ProxyRequest(w http.ResponseWriter, r *http.Request) {

	uri := t.URL + r.RequestURI
	fmt.Println(r.Method + ": " + uri)

	t.MethodHandler(r)

	remote_request, err := t.CreateRemoteRequest(r, uri)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	t.CopyHeader(r.Header, &remote_request.Header)

	resp, err := t.Query(remote_request)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := t.ReadBody(resp)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	destination_header := w.Header()
	t.CopyHeader(resp.Header, &destination_header)
	destination_header.Add("Requested-Host", remote_request.Host)
	w.Write(body)
}

func (t *Target) MethodHandler(r *http.Request) {
	if r.Method == "POST" {
		fmt.Printf("Method is POST:")
	}
}

func (t *Target) CreateRemoteRequest(r *http.Request, uri string) (*http.Request, error) {
	rr, err := http.NewRequest(r.Method, uri, r.Body)
	if err != nil {
		return nil, err
	}
	return rr, nil
}

func (t *Target) Query(r *http.Request) (*http.Response, error) {
	// Create a client and query the target
	var transport http.Transport
	resp, err := transport.RoundTrip(r)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Resp-Headers: %v\n", resp.Header)

	return resp, nil
}

func (t *Target) ReadBody(r *http.Response) ([]byte, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (t *Target) CopyHeader(source http.Header, dest *http.Header) {
	for n, v := range source {
		for _, vv := range v {
			dest.Add(n, vv)
		}
	}
}
