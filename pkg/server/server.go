package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/xegea/webhook_client/pkg/config"
)

type Server struct {
	Config config.Config
}

func NewServer(cfg config.Config) Server {
	svr := Server{
		Config: cfg,
	}
	return svr
}

type Request struct {
	Url     string          `json:"url,omitempty"`
	Host    string          `json:"host,omitempty"`
	Method  string          `json:"method,omitempty"`
	Body    any             `json:"body,omitempty"`
	Headers json.RawMessage `json:"headers,omitempty"`
}

func (s Server) Start(url string) error {

	req := struct {
		Url string
	}{}
	req.Url = url
	json_data, err := json.Marshal(req)
	if err != nil {
		return errors.New("failed to Marshall request")
	}

	b, err := doRequest(s.Config.ServerUrl+"/url", "POST", bytes.NewBuffer(json_data))
	if err != nil {
		return err
	}

	res := struct {
		Token *string `json:"token"`
	}{}
	if err := json.Unmarshal(b, &res); err != nil {
		return fmt.Errorf("failed to unmarshall 'token': %v", err)
	}

	fmt.Println("Congratulations!!!")
	fmt.Printf("You have access to: %s\n", url+"/<your_path>")
	fmt.Printf("using the next url: %s\n\n\n", strings.Join([]string{s.Config.ServerUrl, *res.Token, "<your_path>"}, "/"))

	fmt.Print("waiting for request...")

	ch := make(chan bool)
	go func(ch chan bool, url string) {
		for {
			time.Sleep(1 * time.Second)

			b, err := doRequest(s.Config.ServerUrl+"/pop/"+*res.Token, "GET", bytes.NewBuffer(nil))
			if err != nil {
				fmt.Print(".")
				continue
			}
			var r Request
			if err := json.Unmarshal(b, &r); err != nil {
				log.Printf("failed to unmarshall 'path': %v", err)
				continue
			}

			body, ok := r.Body.([]byte)
			if !ok {
				body = nil
			}

			newUrl := strings.Replace(r.Url, "/"+*res.Token, url, 1)
			fmt.Println()
			log.Println("Start Processing request...", r.Method, newUrl)

			req, err := http.NewRequest(r.Method, newUrl, bytes.NewBuffer(body))
			if err != nil {
				log.Printf("failed to create new request: %v", err)
				continue
			}

			var headers map[string]string
			err = json.Unmarshal([]byte(r.Headers), &headers)
			if err != nil {
				log.Printf("failed to unmarshall headers: %v", err)
			}

			for k, v := range headers {
				req.Header.Add(k, v)
			}

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				log.Printf("failed to process request: %v", err)
				continue
			}

			b, err = io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("failed to read response Body: %v", err)
				continue
			}

			_, err = doRequest(s.Config.ServerUrl+"/resp/"+*res.Token, "POST", bytes.NewBuffer(b))
			if err != nil {
				log.Printf("failed to save response Body: %v", err)
				continue
			}
			//fmt.Println(string(b))

			resp.Body.Close()

			fmt.Println()
			fmt.Print("waiting for new request...")
			//ch <- true
		}
	}(ch, url)
	ended := <-ch
	_ = ended

	return nil
}

func doRequest(url, method string, body *bytes.Buffer) ([]byte, error) {

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header = http.Header{
		"Content-Type": []string{"application/vnd.api+json"},
		"Accept":       []string{"application/vnd.api+json"},
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, errors.New(resp.Status)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read the response: %w", err)
	}

	return b, nil
}
