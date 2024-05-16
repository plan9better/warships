package gameclient

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"warships/httpclient"
	"warships/utils"
)

func printLobby() {
	lobby := httpc.GetLobby()
	if len(lobby) == 0 {
		fmt.Println("Lobby is empty")
	} else {
		fmt.Println("Lobby: ")
		for i := 0; i < len(lobby); i++ {
			fmt.Println(lobby[i].Nick, ": ", lobby[i].GameStatus)
		}
	}
}

func MainMenu(httpcl *httpclient.HttpClient) {
	httpc = httpcl
	nick := utils.PromptString("nick", DEFAULT_NICK)
	desc := utils.PromptString("description", DEFAULT_DESC)

	var cfg httpclient.GameConfig
	cfg.Nick = nick
	cfg.Desc = desc

	choice := ""
	exit := false
	bot := true
	for !exit {
		choice = utils.PromptString("should a bot play for you", "true")
		if choice == "true" {
			bot = true
			exit = true
		} else if choice == "false" {
			bot = false
			exit = true
		}
	}
	exit = false
	for !exit {
		choice = utils.PromptString("human / bot", "bot")
		if choice == "human" || choice == "bot" {
			exit = true
		}
	}
	if choice == "human" {
		for {
			fmt.Println("1. Look at lobby\n2. Get into lobby\n3. Challange a player from the lobby")
			choice = ""
			reader := bufio.NewReader(os.Stdin)
			text, _ := reader.ReadBytes('\n')
			choice = strings.TrimSuffix(string(text), "\n")

			switch choice {
			case "1":
				printLobby()
			case "2":
				cfg.Target = ""
				cfg.Wpbot = false
				token, err := httpc.GetAuthToken(&cfg)
				if err != nil {
					fmt.Println("error getting auth token")
				}
				httpc.AuthToken = token
				var status httpclient.GameStatus
				status.GameStatus = "waiting"
				for status.GameStatus == "waiting" {
					printLobby()
					fmt.Println("Waiting for a challange")
					time.Sleep(3 * time.Second)
					fmt.Println("Refreshing wait timer")
					httpc.RefreshWaitSession()
					time.Sleep(3 * time.Second)
					status, _ = httpc.GetGameStatus()
					fmt.Println(status.GameStatus)
					time.Sleep(4 * time.Second)
				}
				StartGame(httpc, bot)
			case "3":
				targetExists := false
				var target string
				for !targetExists {
					fmt.Print("Target name: ")
					text, _ := reader.ReadBytes('\n')
					target = (strings.TrimSuffix(string(text), "\n"))
					lobby := httpc.GetLobby()

					for i := 0; i < len(lobby); i++ {
						if lobby[i].Nick == target {
							targetExists = true
							break
						}
					}
					if !targetExists {
						fmt.Println("That target does not exist, try again")
					} else {
						break
					}
				}
				cfg.Target = target
				token, err := httpc.GetAuthToken(&cfg)
				if err != nil {
					log.Println("Error getting auth token")
				}
				httpc.AuthToken = token

				StartGame(httpc, bot)
			}
		}
	} else {
		cfg.Target = ""
		cfg.Wpbot = true
		token, err := httpc.GetAuthToken(&cfg)
		if err != nil {
			log.Println("Error getting auth token")
		}
		httpc.AuthToken = token
		StartGame(httpc, bot)
	}
}
