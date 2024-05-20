package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const MAX_TRIES = 5

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

	tries := 1
	var body []byte
	isCritical := false
	for !isCritical {
		// log.Println("Making a request to ", endpoint)
		resp, err := c.Client.Do(req)
		if err != nil {
			log.Printf("Error sending a get request to %s\n", endpoint)
			log.Printf("Error: %s\n", err)
			return err
		}
		defer resp.Body.Close()
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading response body from request to %s\n", endpoint)
			log.Printf("Error: %s\n", err)
			return err
		}
		if resp.StatusCode != 200 {
			isCritical = handleHTTPCodes(resp.StatusCode, body, tries, endpoint)
			tries++
		} else {
			err = json.Unmarshal(body, &v)
			if err != nil {
				log.Printf("Error unmarshaling JSON response: %s\n", err)
				return err
			}
			break
		}
	}

	return nil

}

func handleHTTPCodes(code int, body []byte, tries int, endpoint string) bool {
	log.Printf("_______________________________________\n")
	log.Printf("HTTP Code %d while requesting %s\n", code, endpoint)
	if string(body) == "" {
		log.Println("(Response body is empty)")
	} else {
		log.Println(string(body))
	}

	isCritical := false
	switch code {
	case 400:
		log.Printf("Bad request\n")
		isCritical = true
	case 404:
		log.Printf("Not found")
		isCritical = true
	case 401:
		log.Printf("No auth")
		isCritical = true
	// case 503:
	// 	log.Println("tries: ", tries)
	// 	isCritical = false
	case 429:
		// Too many requests
		time.Sleep(5 * time.Second)
		isCritical = false
	default:
		log.Println("Unexpexted error, retrying: ", tries)
		isCritical = false
	}
	if tries >= MAX_TRIES {
		log.Fatal("Failed after ", MAX_TRIES, " tries. Exiting")
		isCritical = true
	}
	return isCritical
}

func (c *HttpClient) GetGameStatus() (GameStatus, error) {
	var status GameStatus
	err := c.makeRequest("game", &status, "GET", nil)
	tryCounter := 1
	for err != nil && tryCounter < 5 {
		log.Printf("Error getting game status: %s, retrying %d time\n", err, tryCounter)
		err = c.makeRequest("game", &status, "GET", nil)
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
	time.Sleep(3 * time.Second)
	var desc Desc
	tryCounter := 1
	for desc.Opp_Desc == "" && tryCounter < 5 {
		err := c.makeRequest("game/desc", &desc, "GET", nil)
		if err != nil {
			log.Println("Error getting description: ", err)
			return desc, err
		}
		tryCounter++
		time.Sleep(3 * time.Second)

	}
	return desc, nil
}

type Player struct {
	GameStatus string `json:"game_status"`
	Nick       string `json:"nick"`
}

func (c *HttpClient) GetLobby() []Player {
	var players []Player
	err := c.makeRequest("lobby", &players, "GET", nil)
	if err != nil {
		log.Println("Error requesting lobby")
	}
	return players
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

	tries := 1
	isCritical := false
	var resp *http.Response
	for !isCritical {
		req.Header.Set("Content-Type", "application/json")
		resp, err = c.Client.Do(req)
		if err != nil {
			log.Printf("Error sending request to auth\n")
		}
		if resp.StatusCode != 200 {
			isCritical = handleHTTPCodes(resp.StatusCode, nil, tries, "auth")
		} else {
			break
		}
	}
	return resp.Header.Get("X-Auth-Token"), nil

}

func (c *HttpClient) RefreshWaitSession() {
	req, err := http.NewRequest("GET", "https://go-pjatk-server.fly.dev/api/game/refresh", nil)
	if err != nil {
		log.Println("Error creating a request to refresh", err)
	}
	req.Header.Add("X-Auth-Token", c.AuthToken)
	tries := 1
	isCritical := false
	var resp *http.Response
	for !isCritical {
		resp, err = c.Client.Do(req)
		if err != nil {
			log.Printf("Error sending request to refresh\n")
		}
		if resp.StatusCode != 200 {
			isCritical = handleHTTPCodes(resp.StatusCode, nil, tries, "auth")
		} else {
			break
		}
	}
}

func (c *HttpClient) GetGameBoard() []string {
	type board struct {
		Board []string `json:"board"`
	}
	var brd board
	err := c.makeRequest("game/board", &brd, "GET", nil)
	if err != nil {
		log.Println("Error fetching game board")
		return brd.Board
	}
	return brd.Board
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
	c.Client.Do(req)
}
