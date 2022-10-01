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

	fmt.Println("use the next url path template: ")
	fmt.Println(strings.Join([]string{s.Config.ServerUrl, *res.Token, "<your_path_here>"}, "/"))

	for {
		time.Sleep(2 * time.Second)

		b, err := doRequest(s.Config.ServerUrl+"/pop/"+*res.Token, "GET", bytes.NewBuffer(nil))
		if err != nil {
			log.Printf("no request to process...")
			continue
		}
		res := struct {
			Path *string `json:"path"`
		}{}
		if err := json.Unmarshal(b, &res); err != nil {
			log.Printf("failed to unmarshall 'path': %v", err)
			continue
		}

		doRequest(url+"/"+*res.Path, "GET", bytes.NewBuffer(json_data))

		// TODO insert response:token into redis
		break
	}

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
