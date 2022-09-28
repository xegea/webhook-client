package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

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

	b, err := doRequest(s.Config.ServerUrl+"/token", "POST", bytes.NewBuffer(json_data))
	if err != nil {
		return err
	}

	res := struct {
		Token *string `json:"token"`
	}{}
	if err := json.Unmarshal(b, &res); err != nil {
		return fmt.Errorf("failed to unmarshall: %v", err)
	}

	fmt.Println(*res.Token)

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
