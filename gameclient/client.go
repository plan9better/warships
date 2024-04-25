package gameclient

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"warships/httpclient"
	"warships/utils"

	gui "github.com/grupawp/warships-lightgui/v2"
)

const DEFAULT_NICK = "Patryk"
const DEFAULT_DESC = "Majtek"
const DEFAULT_TARGET = ""
const DEFAULT_WPBOT = "true"

var reader *bufio.Reader = bufio.NewReader(os.Stdin)
var cfg *httpclient.GameConfig
var board *gui.Board
var httpc *httpclient.HttpClient

func createHttpc() {
	httpc = &httpclient.HttpClient{
		Client: &http.Client{Timeout: time.Second * 20},
	}
}
func createCfg() {
	nick := utils.PromptString("nick", DEFAULT_NICK)
	desc := utils.PromptString("description", DEFAULT_DESC)
	// target := utils.PromptString("target", DEFAULT_TARGET)

	var wpbot string
	for wpbot != "true" && wpbot != "false" {
		wpbot = utils.PromptString("wpbot", DEFAULT_WPBOT)
	}

	var wpbotBool bool
	if wpbot == "true" {
		wpbotBool = true
	} else {
		wpbotBool = false
	}

	cfg = &httpclient.GameConfig{
		Nick: nick,
		Desc: desc,
		// Target: target,
		Wpbot: wpbotBool,
	}
}

func auth() {
	auth, ok := httpc.GetAuthToken(cfg)
	if ok != 200 {
		log.Println("Invalid response auth token: ", ok)
		if ok == 400 {
			log.Println("Bad request")
		}
	}
	httpc.AuthToken = auth
}

func fire() (string, string) {
	valid := false
	var toFire string

	for !valid {
		fmt.Printf("Fire at: ")
		text, _ := reader.ReadBytes('\n')
		toFire = string(text)
		valid = utils.CheckValidCoords(toFire)
	}

	toFire = toFire[:len(toFire)-1]
	isHit, ok := httpc.Fire(toFire)
	if ok != 200 {
		log.Println("Problem firing", ok)
	}
	return isHit, toFire
}
func getShips() []string {
	ships, ok := httpc.GetGameBoard()
	if ok != 200 {
		log.Println("Error getting game board: ", ok)
	}
	return ships
}

func fireUpdate() {
	isHit, toFire := fire()
	switch isHit {
	case "hit":
		board.Set(gui.Right, toFire, gui.Hit)
	case "miss":
		board.Set(gui.Right, toFire, gui.Miss)
	case "sunk":
		board.Set(gui.Right, toFire, gui.Ship)
	}
}

func checkGameStatus() httpclient.GameStatus {
	status := httpc.GetGameStatus()

	return status
}

func setup() {
	createCfg()
	createHttpc()
	board = gui.New(gui.NewConfig())
}

func printInfo(desc httpclient.Desc, status httpclient.GameStatus) {
	fmt.Println("status:\t", status.GameStatus)
	fmt.Println("Opponent:\t", status.Opponent)
	fmt.Println("Opponent desc:\t", desc.Opp_Desc)
	fmt.Println("my desc:\t", desc.Desc)
}

func oppShotHandler(status httpclient.GameStatus, ships []string) {
	for _, shot := range status.OpponentShots {
		enemyShotHit := gui.Miss
		for _, ship := range ships {
			if shot == ship {
				enemyShotHit = gui.Hit
				board.Set(gui.Left, shot, enemyShotHit)
				break
			}
		}
		board.Set(gui.Left, shot, enemyShotHit)
	}
}

func StartGame() {
	setup()
	auth()
	fmt.Println("Game started")
	time.Sleep(2 * time.Second)
	desc := httpc.GetDesc()
	ships := getShips()
	board.Import(ships)
	for {
		time.Sleep(time.Second)

		status := checkGameStatus()
		if status.GameStatus == "ended" {
			fmt.Println("Game ended")
			break
		}
		oppShotHandler(status, ships)
		board.Display()
		printInfo(desc, status)
		status = checkGameStatus()
		for !status.ShouldFire {
			time.Sleep(time.Second)
			status = checkGameStatus()
		}
		fireUpdate()
	}
}
