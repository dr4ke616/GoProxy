package server_test

import (
	"github.com/dr4ke616/GoProxy/server"
	"log"
	"net/http"
	"net/http/httptest"

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

type testHandler func(w http.ResponseWriter, req *http.Request)

func getTestHandler(code int) testHandler {
	return func(w http.ResponseWriter, req *http.Request) {
		header := w.Header()
		header.Add("Content-Type", "text/plain; charset=utf-8")

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

			It("Should contain three routing option", func() {
				Expect(len(p.RoutingOptions)).To(Equal(3))
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
	})

	It("Should not error", func() {
		Expect(err).NotTo(HaveOccurred())
	})
})
