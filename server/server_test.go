package server_test

import (
	"encoding/json"
	"github.com/dr4ke616/GoProxy/server"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func loadProxy(config string) (*server.Proxy, error) {
	var err error
	proxy := server.Proxy{}

	err = server.LoadFromConfig(&proxy, config)
	if err != nil {
		return nil, err
	}

	return &proxy, nil
}

var rawData string

type testHandler func(w http.ResponseWriter, req *http.Request)

func getTestHandler(code int) testHandler {
	return func(w http.ResponseWriter, req *http.Request) {
		header := w.Header()
		header.Add("Content-Type", "text/plain; charset=utf-8")

		b, err := ioutil.ReadAll(req.Body)
		if err != nil {
			panic(err)
		}
		rawData = string(b)

		w.WriteHeader(code)
		w.Write(nil)
	}
}

func startTestServer() {
	go func() {

		mux := http.NewServeMux()
		mux.HandleFunc("/", getTestHandler(200))
		mux.HandleFunc("/testendpoint1/", getTestHandler(200))
		mux.HandleFunc("/testendpoint2/", getTestHandler(200))
		mux.HandleFunc("/testendpoint3/", getTestHandler(200))
		mux.HandleFunc("/testendpoint4/query", getTestHandler(200))
		mux.HandleFunc("/testendpoint5/query", getTestHandler(200))
		mux.HandleFunc("/doesnt/exist", getTestHandler(404))

		log.Println(http.ListenAndServe("localhost:14200", mux))
	}()
}

var _ = Describe("Server", func() {

	server.DEVIL = true
	startTestServer()

	p, err := loadProxy("server_test_config.json")

	Describe("Testing GoProxy", func() {
		Context("When Proxy started", func() {

			It("Should populate the fields correctly", func() {
				Expect(p.LogFile).To(Equal("goproxy.log"))
				Expect(p.ListeningPort).To(Equal("9090"))
				Expect(p.TargetUrl).To(Equal("http://localhost:14200"))
			})

			req, err := http.NewRequest("GET", p.TargetUrl, nil)
			w := httptest.NewRecorder()
			p.ServeHTTP(w, req)

			It("Should connect to the target URL", func() {
				Expect(w.Code).To(Equal(200))
			})

			It("Should contain Content-Type: text/plain; charset=utf-8 header", func() {
				contentType := w.Header().Get("Content-Type")
				Expect(contentType).To(Equal("text/plain; charset=utf-8"))
			})

			It("Should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("Now we validate our routing options", func() {

			It("Should contain five routing option", func() {
				Expect(len(p.RoutingOptions)).To(Equal(5))
			})
		})

		Context("Make sure response codes work", func() {

			req, err := http.NewRequest("GET", p.TargetUrl+"/doesnt/exist", nil)
			req.RequestURI = "/doesnt/exist"
			w := httptest.NewRecorder()
			p.ServeHTTP(w, req)

			It("Should connect to the target URL", func() {
				Expect(w.Code).To(Equal(404))
			})

			It("Should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("On the 1st routing option", func() {

			req, err := http.NewRequest("GET", p.TargetUrl+p.RoutingOptions[0].URI, nil)
			req.RequestURI = p.RoutingOptions[0].URI
			w := httptest.NewRecorder()
			p.ServeHTTP(w, req)

			It("Should accept connections from custom endpoint", func() {
				Expect(w.Code).To(Equal(200))
			})

			It("Should contain Content-Type: text/plain; charset=utf-8 header", func() {
				contentType := w.Header().Get("Content-Type")
				Expect(contentType).To(Equal("text/plain; charset=utf-8"))
			})

			It("Should contain the POST method type", func() {
				method := req.Method
				Expect(method).To(Equal("POST"))
			})

			It("Should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("On the 2nd routing option", func() {

			req, err := http.NewRequest("GET", p.TargetUrl+p.RoutingOptions[1].URI, nil)
			req.RequestURI = p.RoutingOptions[1].URI
			w := httptest.NewRecorder()
			p.ServeHTTP(w, req)

			It("Should accept connections from custom endpoint", func() {
				Expect(w.Code).To(Equal(200))
			})

			It("Checking the headers. This context will append to the header key", func() {
				contentType := w.Header().Get("Content-Type")
				Expect(contentType).To(Equal("text/plain; charset=utf-8, application/json, text/plain"))
			})

			It("Should contain the GET method type", func() {
				method := req.Method
				Expect(method).To(Equal("GET"))
			})

			It("Should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("On the 3nd routing option", func() {

			req, err := http.NewRequest("POST", p.TargetUrl+p.RoutingOptions[2].URI, nil)
			req.RequestURI = p.RoutingOptions[2].URI
			w := httptest.NewRecorder()
			p.ServeHTTP(w, req)

			It("Should accept connections from custom endpoint", func() {
				Expect(w.Code).To(Equal(200))
			})

			It("Checking the headers. This context will replace the header key", func() {
				contentType := w.Header().Get("Content-Type")
				Expect(contentType).To(Equal("application/json, text/plain"))
			})

			It("Should contain the GET method type", func() {
				method := req.Method
				Expect(method).To(Equal("GET"))
			})

			It("Should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("On the 4th routing option - test copy params to a JSON body", func() {

			params :=
				"param1=foo&param2=10&param3=false&param4=true" +
					"&copy_string=test1&copy_string=test2" +
					"&copy_int=100&copy_int=150&" +
					"copy_bool=false&copy_bool=true"

			uri := p.RoutingOptions[3].URI + "?" + params
			req, err := http.NewRequest("GET", p.TargetUrl+uri, nil)
			req.RequestURI = uri
			w := httptest.NewRecorder()
			p.ServeHTTP(w, req)

			It("Should copy paramaters flag in config is true", func() {
				Expect(p.RoutingOptions[3].CopyParamaters).To(BeTrue())
			})

			It("Should accept connections from custom endpoint", func() {
				Expect(w.Code).To(Equal(200))
			})

			It("Checking the headers. This context will replace the header key", func() {
				contentType := w.Header().Get("Content-Type")
				Expect(contentType).To(Equal("application/json"))
			})

			It("Should contain the PATCH method type", func() {
				method := req.Method
				Expect(method).To(Equal("PATCH"))
			})

			type JSONBody struct {
				Param1     string    `json:"param1"`
				Param2     int       `json:"param2"`
				Param3     bool      `json:"param3"`
				Param4     bool      `json:"param4"`
				CopyString [2]string `json:"copy_string"`
				CopyInt    [2]int    `json:"copy_int"`
				CopyBool   [2]bool   `json:"copy_bool"`
			}
			jsonBody := &JSONBody{}
			err = json.Unmarshal([]byte(rawData), &jsonBody)

			It("Should contain a JSON encoded body", func() {
				Expect(jsonBody.Param1).To(Equal("foo"))
				Expect(jsonBody.Param2).To(Equal(10))
				Expect(jsonBody.Param3).To(BeFalse())
				Expect(jsonBody.Param4).To(BeTrue())
				Expect(jsonBody.CopyString).To(Equal([2]string{"test1", "test2"}))
				Expect(jsonBody.CopyInt).To(Equal([2]int{100, 150}))
				Expect(jsonBody.CopyBool).To(Equal([2]bool{false, true}))
			})

			It("Should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("On the 5th routing option - test copy params to a X-WWW-Form body", func() {

			params :=
				"param1=foo&param2=10&param3=false&param4=true" +
					"&copy_string=test1&copy_string=test2" +
					"&copy_int=100&copy_int=150&" +
					"copy_bool=false&copy_bool=true"

			uri := p.RoutingOptions[4].URI + "?" + params
			req, err := http.NewRequest("GET", p.TargetUrl+uri, nil)
			req.RequestURI = uri
			w := httptest.NewRecorder()
			p.ServeHTTP(w, req)

			It("Should copy paramaters flag in config is true", func() {
				Expect(p.RoutingOptions[3].CopyParamaters).To(BeTrue())
			})

			It("Should accept connections from custom endpoint", func() {
				Expect(w.Code).To(Equal(200))
			})

			It("Checking the headers. This context will replace the header key", func() {
				contentType := w.Header().Get("Content-Type")
				Expect(contentType).To(Equal("application/x-www-form-urlencoded"))
			})

			It("Should contain the POST method type", func() {
				method := req.Method
				Expect(method).To(Equal("POST"))
			})

			values, err := url.ParseQuery(rawData)
			It("Should contain a URL encoded body", func() {
				Expect(values["param1"]).To(Equal([]string{"foo"}))
				Expect(values["param2"]).To(Equal([]string{"10"}))
				Expect(values["param3"]).To(Equal([]string{"false"}))
				Expect(values["param4"]).To(Equal([]string{"true"}))
				Expect(values["copy_string"]).To(Equal([]string{"test1", "test2"}))
				Expect(values["copy_int"]).To(Equal([]string{"100", "150"}))
				Expect(values["copy_bool"]).To(Equal([]string{"false", "true"}))
			})

			It("Should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	It("Should not error", func() {
		Expect(err).NotTo(HaveOccurred())
	})
})
