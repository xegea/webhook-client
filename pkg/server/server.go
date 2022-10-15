package server

import (
	"bytes"
	"context"
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
	Id      string          `json:"id,omitempty"`
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
		return errors.New("failed to marshal request")
	}

	b, err := requestServerWithContext(s.Config.ServerUrl+"/url", "POST", bytes.NewBuffer(json_data))
	if err != nil {
		return err
	}

	res := struct {
		Token *string `json:"token"`
	}{}
	if err := json.Unmarshal(b, &res); err != nil {
		return fmt.Errorf("failed to unmarshal 'token': %v", err)
	}

	fmt.Println("Congrats!")
	fmt.Printf("You have access to: %s\n", url+"/<your_path>")
	fmt.Printf("using the next url: %s\n\n\n", strings.Join([]string{s.Config.ServerUrl, *res.Token, "<your_path>"}, "/"))

	fmt.Print("waiting for requests...")

	ch := make(chan bool)
	go func(ch chan bool, url string) {
		for {
			time.Sleep(100 * time.Millisecond)

			b, err := requestServerWithContext(s.Config.ServerUrl+"/pop/"+*res.Token, "GET", bytes.NewBuffer(nil))
			if err != nil {
				fmt.Print(".")
				continue
			}
			var rs []Request
			if err := json.Unmarshal(b, &rs); err != nil {
				log.Printf("failed to unmarshal 'path': %v", err)
				continue
			}

			for _, r := range rs {

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
					log.Printf("failed to unmarshal headers: %v", err)
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

				response := struct {
					Body    []byte              `json:"body"`
					Headers map[string][]string `json:"headers"`
				}{}

				response.Body = b
				response.Headers = resp.Header

				b, err = json.Marshal(response)
				if err != nil {
					log.Printf("failed to marshal response: %v", err)
					continue
				}

				_, err = requestServerWithContext(s.Config.ServerUrl+"/resp/"+*res.Token+"~"+r.Id, "POST", bytes.NewBuffer(b))
				if err != nil {
					log.Printf("failed to save response Body: %v", err)
					continue
				}
				//fmt.Println(string(b))

				resp.Body.Close()
			}

			//ch <- true
		}
	}(ch, url)
	ended := <-ch
	_ = ended

	return nil
}

func requestServerWithContext(url, method string, body *bytes.Buffer) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("requestServerWithContext: failed creating request: %w", err)
	}

	req.Header = http.Header{
		"Content-Type": []string{"application/vnd.api+json"},
		"Accept":       []string{"application/vnd.api+json"},
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("requestServerWithContext: failed executing request %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, errors.New(res.Status)
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("requestServerWithContext: failed reading response %w", err)
	}

	return b, nil
}
