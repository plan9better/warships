package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type HttpClient struct {
	Client    *http.Client
	AuthToken string
}

func (c *HttpClient) GetGameStatus() ([]byte, int) {

	req, err := http.NewRequest("GET", "https://go-pjatk-server.fly.dev/api/game", nil)
	if err != nil {
		log.Println("Error creating a request to check game status: ", err)
	}
	req.Header.Add("X-Auth-Token", c.AuthToken)

	resp, err := c.Client.Do(req)
	if err != nil {
		log.Println("Error sending request while checking game status")
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading body while checking game status: ", err)
	}
	return body, resp.StatusCode
}

type GameConfig struct {
	Wpbot  bool   `json:"wpbot"`
	Desc   string `json:"desc"`
	Nick   string `json:"nick"`
	Coords []byte `json:"coords"`
	Target string `json:"target_nick"`
}

func (c *HttpClient) GetAuthToken(cfg *GameConfig) (string, int) {

	bm, err := json.Marshal(cfg)
	if err != nil {
		log.Fatal("Error marshaling", err)
	}

	req, err := http.NewRequest("POST", "https://go-pjatk-server.fly.dev/api/game", bytes.NewReader(bm))
	if err != nil {
		log.Fatal("Error creating a request", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.Client.Do(req)
	if err != nil {
		log.Fatal("Error requesting /game", err)
	}
	return resp.Header.Get("X-Auth-Token"), resp.StatusCode

}

func (c *HttpClient) GetGameBoard() ([]byte, int) {
	req, err := http.NewRequest("GET", "https://go-pjatk-server.fly.dev/api/game/board", nil)
	if err != nil {
		log.Println("Error creating request while getting game board", err)
	}
	req.Header.Add("X-Auth-Token", c.AuthToken)

	resp, err := c.Client.Do(req)
	if err != nil {
		log.Println("Error requesting game board ", err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body while getting game board ", err)
	}

	return body, resp.StatusCode
}

func (c *HttpClient) Fire(x byte, y byte) ([]byte, int) {
	type coord struct {
		Coord string `json:"coord"`
	}

	var toFire string
	toFire = string(x)
	toFire += string(y)

	crd := &coord{
		Coord: toFire,
	}

	crdm, err := json.Marshal(crd)
	if err != nil {
		log.Println("Error marshaling fire coords")
	}

	fmt.Println("COOOOORD: ", string(crdm))
	req, err := http.NewRequest("POST", "https://go-pjatk-server.fly.dev/api/game/fire", bytes.NewReader(crdm))
	if err != nil {
		log.Println("Error creating request while firing", err)
	}
	req.Header.Add("X-Auth-Token", c.AuthToken)

	resp, err := c.Client.Do(req)
	if err != nil {
		log.Println("Error firing", err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body while firing", err)
	}

	return body, resp.StatusCode

}
