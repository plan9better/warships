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

type GameStatus struct {
	Nick           string   `json:"nick"`
	GameStatus     string   `json:"game_status"`
	LastGameStatus string   `json:"last_game_status"`
	Opponent       string   `json:"opponent"`
	OpponentShots  []string `json:"opp_shots"`
	Timer          int      `json:"timer"`
	ShouldFire     bool     `json:"should_fire"`
}

func (c *HttpClient) makeRequest(endpoint string, v any, method string, payload io.Reader) error {
	address := fmt.Sprintf("https://go-pjatk-server.fly.dev/api/%s", endpoint)
	req, err := http.NewRequest(method, address, payload)
	if err != nil {
		log.Printf("Error making a get request to %s\n", endpoint)
		log.Printf("Error: %s\n", err)
		return err
	}

	req.Header.Add("X-Auth-Token", c.AuthToken)
	resp, err := c.Client.Do(req)
	if err != nil {
		log.Printf("Error sending a get request to %s\n", endpoint)
		log.Printf("Error: %s\n", err)
		return err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body from request to %s\n", endpoint)
		log.Printf("Error: %s\n", err)
		return err
	}
	handleHTTPCodes(resp.StatusCode, body)

	err = json.Unmarshal(body, &v)
	if err != nil {
		log.Printf("Error unmarshaling JSON response\n")
		return err
	}
	handleHTTPCodes(resp.StatusCode, body)
	return nil

}

func handleHTTPCodes(code int, body []byte) {
	log.Printf("HTTP Code %d\n", code)
	switch code {
	case 400:
		log.Printf("Bad request\n")
	case 404:
		log.Printf("Not found")
	case 505:
	}
	log.Println(string(body))
}

func (c *HttpClient) GetGameStatus() (GameStatus, error) {
	var status GameStatus
	err := c.makeRequest("game", &status, "GET", nil)
	tryCounter := 1
	for err != nil && tryCounter < 5 {
		log.Printf("Error getting game status: %s, retrying %d time\n", err, tryCounter)
		return status, err
	}

	return status, nil
}

type GameConfig struct {
	Wpbot  bool   `json:"wpbot"`
	Desc   string `json:"desc"`
	Nick   string `json:"nick"`
	Coords []byte `json:"coords"`
	Target string `json:"target_nick"`
}

type Desc struct {
	Desc     string `json:"desc"`
	Nick     string `json:"nick"`
	Opp_Desc string `json:"opp_desc"`
	Opponent string `json:"opponent"`
}

func (c *HttpClient) GetDesc() (Desc, error) {
	var desc Desc
	err := c.makeRequest("game/desc", &desc, "GET", nil)
	if err != nil {
		log.Println("Error getting description: ", err)
		return desc, err
	}
	return desc, nil
}

func (c *HttpClient) GetAuthToken(cfg *GameConfig) (string, error) {
	bm, err := json.Marshal(cfg)
	if err != nil {
		log.Fatal("Error marshaling request for auth token", err)
		return "", err
	}

	req, err := http.NewRequest("POST", "https://go-pjatk-server.fly.dev/api/game", bytes.NewReader(bm))
	if err != nil {
		log.Println("Error creating a request", err)
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.Client.Do(req)
	tryCounter := 1
	if err != nil && tryCounter < 5 {

		// log.Fatal("Error requesting /game", err)
		// return "", err
	}
	return resp.Header.Get("X-Auth-Token"), nil

}

func (c *HttpClient) GetGameBoard() ([]string, error) {
	// var board []string
	type board struct {
		Board []string `json:"board"`
	}
	var brd board
	err := c.makeRequest("game/board", &brd, "GET", nil)
	if err != nil {
		log.Println("Error fetching game board")
		return brd.Board, err
	}
	return brd.Board, nil
}

func (c *HttpClient) Fire(toFire string) (string, error) {
	type coord struct {
		Coord string `json:"coord"`
	}
	var crd coord
	crd.Coord = toFire

	crdm, err := json.Marshal(crd)
	if err != nil {
		log.Println("Error marshaling fire coords")
	}

	type result struct {
		Result string `json:"result"`
	}
	var res result
	err = c.makeRequest("game/fire", &res, "POST", bytes.NewReader(crdm))
	if err != nil {
		log.Printf("Error firing: %s\n", err)
		return res.Result, err
	}
	return res.Result, nil
}

func (c *HttpClient) Abandon() {
	req, _ := http.NewRequest("DELETE", "https://go-pjatk-server.fly.dev/api/game/fire", nil)
	req.Header.Add("X-Auth-Token", c.AuthToken)
	resp, _ := c.Client.Do(req)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
}
