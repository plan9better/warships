package gameclient

import (
	"bufio"
	"fmt"
	"log"
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
	auth, err := httpc.GetAuthToken(cfg)
	tryCounter := 1

	if err != nil && tryCounter < 3 {
		log.Println("Invalid response auth token: ", err)
		log.Println("Retrying...")
		auth, err = httpc.GetAuthToken(cfg)
		tryCounter++
	}
	if err != nil {
		log.Println("Failed to authenticate... exiting")
		return
	}
	httpc.AuthToken = auth
}

func fire() (string, string, error) {
	valid := false
	var toFire string

	for !valid {
		fmt.Printf("Fire at: ")
		text, _ := reader.ReadBytes('\n')
		toFire = string(text)
		valid = utils.CheckValidCoords(toFire)
	}

	toFire = toFire[:len(toFire)-1]
	isHit, err := httpc.Fire(toFire)
	if err != nil {
		log.Println("Error firing")
		return "", toFire, err
	}
	return isHit, toFire, err
}

func fireUpdate() {

	isHit, toFire, err := fire()
	tryCounter := 1
	for err != nil && tryCounter < 3 {
		isHit, toFire, err = fire()
		tryCounter++
	}
	if err != nil {
		log.Println("Failed to fire after 3 tries: ", err)
		return
	}
	switch isHit {
	case "hit":
		board.Set(gui.Right, toFire, gui.Hit)
	case "miss":
		board.Set(gui.Right, toFire, gui.Miss)
	case "sunk":
		board.Set(gui.Right, toFire, gui.Ship)
	}
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

func gameDesc() (httpclient.Desc, error) {
	desc, err := httpc.GetDesc()
	tryCounter := 1
	for err != nil && tryCounter < 3 {
		time.Sleep(time.Second)
		desc, err = httpc.GetDesc()
	}
	return desc, err
}

func gameShips() ([]string, error) {
	ships, err := httpc.GetGameBoard()
	tryCounter := 1
	if err != nil && tryCounter < 3 {
		log.Println("Error getting game board: ", err, " retrying...")
		ships, err = httpc.GetGameBoard()
		tryCounter++
	}
	return ships, err
}

func gameStatus() (httpclient.GameStatus, error) {
	status, err := httpc.GetGameStatus()
	tryCounter := 1
	if err != nil && tryCounter < 3 {
		log.Println("Error getting game status...", err, " retrying")
		status, err = httpc.GetGameStatus()
		tryCounter++
	}
	return status, err
}

func StartGame(httpcl *httpclient.HttpClient) {
	httpc = httpcl
	createCfg()
	board = gui.New(gui.NewConfig())
	auth()
	fmt.Println("Game started")

	desc, err := gameDesc()
	if err != nil {
		log.Println("Getting description failed: ", err)
	}

	ships, err := gameShips()
	if err != nil {
		log.Println("Failed to get ships after 3 tries: ", err, " exiting...")
		return
	}

	board.Import(ships)

	for {
		status, err := gameStatus()
		if err != nil {
			log.Println("Failed to get status after 3 tries: ", err, " exiting...")
			return
		}
		if status.GameStatus == "ended" {
			fmt.Println("Game ended")
			break
		}

		// Wait for your turn
		for !status.ShouldFire && status.GameStatus != "ended" {
			time.Sleep(time.Second)
			status, err = gameStatus()
		}
		// Your turn
		fireUpdate()

		oppShotHandler(status, ships)
		board.Display()
		printInfo(desc, status)
	}
}
