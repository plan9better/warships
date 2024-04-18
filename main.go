package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"warships/httpclient"
)

func main() {

	httpc := &httpclient.HttpClient{
		Client: &http.Client{Timeout: time.Second * 10},
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

	// status, ok := httpc.GetGameStatus()
	// if ok != 200 {
	// 	log.Println("Error ", ok)
	// }
	// fmt.Println("Game status: ", string(status))

	// board, ok := httpc.GetGameBoard()
	// if ok != 200 {
	// 	log.Println("slakdjf")
	// }

	reader := bufio.NewReader(os.Stdin)
	for {
		status, ok := httpc.GetGameStatus()
		if ok != 200 {
			log.Println("Error ", ok)
		}
		fmt.Println("Game status: ", string(status))

		board, ok := httpc.GetGameBoard()
		if ok != 200 {
			log.Println("Error getting game board: ", ok)
		}
		fmt.Println("Game board: ", string(board))
		time.Sleep(5 * time.Second)

		text, _ := reader.ReadBytes('\n')

		body, ok := httpc.Fire(text[0], text[1])
		if ok != 200 {
			log.Println("Problem firing", ok)
		}
		fmt.Println(string(body))
	}

}
