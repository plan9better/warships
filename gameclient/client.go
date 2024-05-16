package gameclient

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
	"warships/httpclient"
	"warships/utils"

	gui "github.com/rrekaf/warships-lightgui"
)

var boardcoord map[Coord]string

const DEFAULT_NICK = "Patryk"
const DEFAULT_DESC = "Majtek"
const DEFAULT_TARGET = ""

var reader *bufio.Reader = bufio.NewReader(os.Stdin)

var board *gui.Board
var httpc *httpclient.HttpClient

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

func fireUpdate() (string, string) {
	isHit, toFire, err := fire()
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
		board.Set(gui.Right, toFire, gui.Miss)
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
				boardcoord[v] = "empty"
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
			} else if boardcoord[crd] == "empty" {
				board.Set(gui.Right, coordToStr(crd), gui.Miss)
			}
		}
	}
	board.Display()
}

func StartGame(httpcl *httpclient.HttpClient) {
	httpc = httpcl
	boardcoord = make(map[Coord]string)

	board = gui.New(gui.NewConfig())

	ships, err := gameShips()
	if err != nil {
		log.Println("Failed to get ships after 3 tries: ", err, " exiting...")
		return
	}

	board.Import(ships)
	desc, err := httpc.GetDesc()
	if err != nil {
		log.Println("Getting description failed: ", err)
	}

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
			if err != nil {
				log.Println("error checking turn", err)
			}
		}
		oppShotHandler(status, ships)
		displayBoard()

		printInfo(desc, status)
		// Your turn
		handlePlayerShot()
		// for i := 'A'; i <= 'J'; i++ {
		// 	for j := 1; j <= 10; j++ {
		// 		if boardcoord[Coord{X: int(i), Y: j}] == "sunk" {
		// 			fmt.Println("i: ", i, "j: ", j, boardcoord[Coord{X: int(i), Y: j}])
		// 		}
		// 	}
		// }
		displayBoard()
		printInfo(desc, status)

	}
}
