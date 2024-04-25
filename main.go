package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"warships/httpclient"

	gui "github.com/grupawp/warships-lightgui/v2"
)

func main() {

	// Logger
	f, err := os.OpenFile("text.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	logger := log.New(f, "", log.LstdFlags)
	logger.Println("~~~~~~~~New game~~~~~~~~")

	// GUI
	board := gui.New(gui.NewConfig())

	// HTTP CLIENT
	httpc := &httpclient.HttpClient{
		Client: &http.Client{Timeout: time.Second * 20},
	}
	cfg := &httpclient.GameConfig{
		Nick: "Patryk",
		Desc: "Majtek",
		// Target: "",
		Wpbot: true,
	}

	auth, ok := httpc.GetAuthToken(cfg)
	if ok != 200 {
		log.Fatal("Invalid response auth token", ok)
	}
	httpc.AuthToken = auth

	reader := bufio.NewReader(os.Stdin)
	for {
		boardSlice, ok := httpc.GetGameBoard()
		if ok != 200 {
			log.Println("Error getting game board: ", ok)
		}

		status, ok := httpc.GetGameStatus()
		if ok != 200 {
			log.Println("Error ", ok)
		}
		if status.GameStatus == "ended" {
			break
		}

		// check if enemy hit your ship
		opShotsLen := len(status.OpponentShots)
		enemyShotHit := gui.Miss
		if opShotsLen > 0 {
			enemyShot := status.OpponentShots[opShotsLen-1]
			for _, i := range boardSlice {
				if enemyShot == i {
					enemyShotHit = gui.Hit
				}
			}

			board.Set(gui.Left, status.OpponentShots[opShotsLen-1], enemyShotHit)
		}

		board.Import(boardSlice)
		board.Display()
		fmt.Println("status: ", status.GameStatus)

		fmt.Printf("Enter coordinates to fire to: ")
		text, _ := reader.ReadBytes('\n')
		toFire := string(text)
		toFire = toFire[:len(toFire)-1]

		isHit, ok := httpc.Fire(toFire)
		if ok != 200 {
			log.Println("Problem firing", ok)
			logger.Println("problem firing: ", ok)
		}
		logger.Println(toFire, ": ", isHit)
		switch isHit {
		case "hit":
			board.Set(gui.Right, toFire, gui.Hit)
		case "miss":
			board.Set(gui.Right, toFire, gui.Miss)
		case "sunk":
			board.Set(gui.Right, toFire, gui.Ship)
		default:
			logger.Println("wrong coords?")
		}

	}

}
