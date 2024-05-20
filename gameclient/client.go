package gameclient

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"os"
	"strconv"
	"time"
	"warships/httpclient"
	"warships/utils"

	gui "github.com/rrekaf/warships-lightgui"
)

var boardcoord map[Coord]string
var bot bool

const DEFAULT_NICK = "Patryk"
const DEFAULT_DESC = "Majtek"
const DEFAULT_TARGET = ""

var reader *bufio.Reader = bufio.NewReader(os.Stdin)

var board *gui.Board
var httpc *httpclient.HttpClient

var botPossibleShots []Coord

func fire() (string, string, error) {
	var toFire string
	var isHit string

	// Human choses shot
	if !bot {
		time.Sleep(1 * time.Second)
		valid := false

		for !valid {
			fmt.Printf("Fire at: ")
			text, _ := reader.ReadBytes('\n')
			toFire = string(text)
			valid = utils.CheckValidCoords(toFire)
		}

		toFire = toFire[:len(toFire)-1]
		isHit, _ := httpc.Fire(toFire)
		return isHit, toFire, nil

	}

	// Bot choses shot
	time.Sleep(time.Second / 2)
	if len(botPossibleShots) == 0 {
		roll := true
		var toF []byte
		for roll {
			toF = []byte{'A', '1'}

			r1 := rand.IntN(10)
			r2 := rand.IntN(10)
			toF[0] += byte(r1)
			toF[1] += byte(r2)
			if boardcoord[strToCoord(string(toF))] == "" {
				roll = false
			}
		}

		toFire = string(toF)
		isHit, _ = httpc.Fire(toFire)

		if isHit == "hit" {
			botPossibleShots = FindToShoot(strToCoord(toFire))
		}

	} else {
		toFire = coordToStr(botPossibleShots[0])
		isHit, _ = httpc.Fire(toFire)
		fmt.Println("isHit", isHit)
		time.Sleep(2 * time.Second)
		// if err != nil {
		// 	log.Println("Error firing")
		// 	time.Sleep(5 * time.Second)
		// 	return "", toFire, err
		// }
		if isHit == "miss" {
			botPossibleShots = append(botPossibleShots[:0], botPossibleShots[1:]...)
		}
		if isHit == "sunk" {
			botPossibleShots = []Coord{}
		}
		if isHit == "hit" {
			// remove the hit and handle the rest
			botPossibleShots = append(botPossibleShots[:0], botPossibleShots[1:]...)
		}
	}
	fmt.Println(isHit, toFire)
	time.Sleep(2 * time.Second)
	return isHit, toFire, nil
}

func fireUpdate() (string, string) {
	isHit, toFire, err := fire()
	fmt.Println("fireUpdate", isHit, toFire)
	time.Sleep(3 * time.Second)
	tryCounter := 1
	for err != nil && tryCounter < 3 {
		isHit, toFire, err = fire()
		tryCounter++
	}
	if err != nil {
		log.Println("Failed to fire after 3 tries: ", err)
		return "", ""
	}
	switch isHit {
	case "hit":
		board.Set(gui.Right, toFire, gui.Hit)
	case "miss":
		if boardcoord[strToCoord(toFire)] == "hit" {
			break
		} else {
			board.Set(gui.Right, toFire, gui.Miss)
		}
	case "sunk":
		board.Set(gui.Right, toFire, gui.Ship)
	}

	// add to boardCoord struct
	boardcoord[strToCoord(toFire)] = isHit

	return isHit, toFire
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

func gameShips() ([]string, error) {
	ships, err := httpc.GetGameBoard()
	tryCounter := 1
	if err != nil && tryCounter < 3 {
		log.Println("Error getting game board: ", err, " retrying...")
		time.Sleep(time.Second)
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
func strToCoord(str string) Coord {
	c := []byte(str)
	var coord Coord
	coord.X = int(c[0])
	if len(c) == 3 {
		coord.Y = 10
	} else {
		coord.Y = int(str[1] - '0')
	}
	return coord
}
func handlePlayerShot() {
	effect, coordStr := fireUpdate()
	coord := strToCoord(coordStr)
	var v Coord

	tries := 0
	for effect == "sunk" && tries < 5 {
		adj := FindAdjacent(coord)
		for _, v = range adj {
			if boardcoord[v] == "hit" || boardcoord[v] == "sunk" {
				boardcoord[v] = "sunk"
				coord = v
			} else {
				boardcoord[v] = "miss"
			}
		}
		tries++
	}
	displayBoard()
}

func coordToStr(coord Coord) string {
	res := string(rune(coord.X))
	res += strconv.Itoa(coord.Y)
	return res
}

func displayBoard() {
	var crd Coord
	for i := 'A'; i <= 'J'; i++ {
		for j := 1; j <= 10; j++ {
			crd = Coord{X: int(i), Y: j}

			if boardcoord[crd] == "sunk" {
				board.Set(gui.Right, coordToStr(crd), gui.Ship)
			} else if boardcoord[crd] == "miss" {
				board.Set(gui.Right, coordToStr(crd), gui.Miss)
			}
		}
	}
	board.Display()
}

func StartGame(httpcl *httpclient.HttpClient, b bool) {
	httpc = httpcl
	bot = b
	boardcoord = make(map[Coord]string)

	board = gui.New(gui.NewConfig())

	ships, err := gameShips()
	if err != nil {
		log.Println("Failed to get ships after 3 tries: ", err, " exiting...")
		return
	}

	// TODO: add option to play continously

	board.Import(ships)
	desc, err := httpc.GetDesc()
	if err != nil {
		log.Println("Getting description failed: ", err)
	}
	type gameJson struct {
		Board map[Coord]string `json:"board"`
		Id    int              `json:"ID"`
	}

	type gamesJson struct {
		Games []gameJson `json:"games"`
	}

	for {
		status, err := gameStatus()
		if err != nil {
			log.Println("Failed to get status after 3 tries: ", err, " exiting...")
			return
		}

		if status.GameStatus == "ended" {
			fmt.Println("Game ended")

			///////
			if bot {

				jsonFile, err := os.Open("log/games.json")
				if err != nil {
					fmt.Println(err)
				}
				jsonBytes, err := io.ReadAll(jsonFile)
				if err != nil {
					log.Println("error reading")
					log.Println(err)
				}

				var games gamesJson
				json.Unmarshal(jsonBytes, &games)
				jsonFile.Close()

				game := gameJson{Board: boardcoord}
				game.Id = games.Games[len(games.Games)-1].Id + 1
				// game.Id = 0
				games.Games = append(games.Games, game)

				jsonBytes, err = json.Marshal(games)
				if err != nil {
					log.Println("error marshaling")
					log.Println(err)
				}

				// fmt.Printf("%+v", games)
				err = os.WriteFile("log/games.json", jsonBytes, 0644)
				if err != nil {
					log.Println(err)
				}
				////////
			}

			break
		}

		// Wait for your turn
		for !status.ShouldFire && status.GameStatus != "ended" {
			time.Sleep(time.Second * 1)
			status, err = gameStatus()
			if err != nil {
				log.Println("error checking turn", err)
			}
		}
		oppShotHandler(status, ships)
		// displayBoard()

		printInfo(desc, status)
		// Your turn
		handlePlayerShot()

		// displayBoard()
		printInfo(desc, status)

	}
}
