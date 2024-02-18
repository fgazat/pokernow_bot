package calc

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"

	"github.com/fgazat/poker/pkg/utils"
)

var users = []UserFromTable{}

type (
	Payment struct {
		Amount    float64
		Recipient *Player
	}

	Player struct {
		Nickname    string
		PaymentInfo string
		Login       string

		Diff   float64
		BuyIn  float64
		BuyOut float64

		Payments []Payment
	}

	UserFromTable struct {
		Nicknames   []string `json:"Nicknames"`
		Login       string   `json:"Login"`
		PaymentInfo string   `json:"PaymentInfo"`
	}
)

func Map(nickname string, login string, usersJsonPath string) error {
	setUsers(usersJsonPath)
	err := checkInput(map[string]string{
		"IN_GAME_NICKNAME": nickname,
		"LOGIN":            login,
	})
	if err != nil {
		return err
	}
	if !strings.HasPrefix(login, "@") {
		login = "@" + login
	}
	found := false
	for i := 0; i < len(users); i++ {
		user := users[i]
		if user.Login == login {
			user.Nicknames = append(user.Nicknames, nickname)
			found = true
			users[i] = user
			break
		}
	}
	if !found {
		return fmt.Errorf("no user found with login: %s, check the input or use `/new` command", login)
	}
	data, err := json.Marshal(&users)
	if err != nil {
		return err
	}
	return os.WriteFile(usersJsonPath, data, 0644)
}

func checkInput(params map[string]string) error {
	var err error
	for name, value := range params {
		if value == "" {
			err = errors.Join(err, fmt.Errorf("no %s provided", name))
		}
	}
	return err
}

func New(nickname string, login string, paymentInfo string, usersJsonPath string) error {
	setUsers(usersJsonPath)
	err := checkInput(map[string]string{
		"IN_GAME_NICKNAME": nickname,
		"LOGIN":            login,
		"PAYMENT_INFO":     paymentInfo,
	})
	if err != nil {
		return err
	}
	if !strings.HasPrefix(login, "@") {
		login = "@" + login
	}
	for _, user := range users {
		if user.Login == login {
			return fmt.Errorf("user with login %s already exists", login)
		}
		if utils.Contains(user.Nicknames, nickname) {
			return fmt.Errorf("this IN_GAME_NICKNAME is already occupied by %s\\. Use another nickname", user.Login)
		}
	}
	users = append(users, UserFromTable{
		Nicknames:   []string{nickname},
		Login:       login,
		PaymentInfo: paymentInfo,
	})
	data, err := json.Marshal(&users)
	if err != nil {
		return err
	}
	return os.WriteFile(usersJsonPath, data, 0644)
}

func init() {
	table := os.Getenv("USERS_JSON_TABLE")
	if table != "" {
		setUsers(table)
	}
}

func setUsers(usersJsonPath string) error {
	data, err := os.ReadFile(usersJsonPath)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, &users); err != nil {
		return err
	}
	return nil
}

func Calcuclate(url string, date string, usersJsonPath string, downloader Downloader) (string, error) {
	var err error
	if len(users) == 0 { // for test purposes
		if err = setUsers(usersJsonPath); err != nil {
			return "", err
		}
	}
	transactions, err := downloader.Download(url)
	if err != nil {
		return "", err
	}
	playersList, err := getPlayersList(transactions)
	if err != nil {
		return "", err
	}
	playersList = processPlayers(playersList)
	comment := getComment(playersList, date)
	return comment, nil
}

func getPlayersList(transactions []TransationInfo) ([]Player, error) {
	playersMap := map[string]Player{}
	missingUsers := []string{}
	for _, transaction := range transactions {
		user := getUserInfo(transaction.PlayerNickname)
		if user == nil {
			missingUsers = append(missingUsers, transaction.PlayerNickname)
			continue
		}
		player, ok := playersMap[user.Login]
		if !ok {
			player = Player{
				Nickname:    transaction.PlayerNickname,
				PaymentInfo: user.PaymentInfo,
				Login:       user.Login,
			}
		}
		player.BuyIn += transaction.BuyIn
		if transaction.Stack != 0 {
			transaction.BuyOut += transaction.Stack
		}
		player.BuyOut += transaction.BuyOut
		player.Diff += float64(transaction.BuyOut - transaction.BuyIn)
		playersMap[user.Login] = player
	}
	if len(missingUsers) != 0 {
		return nil, fmt.Errorf("unknown users found: %s\n\nPlease use `/map` or `/new` commands and relaunch command `/calc`", strings.Join(missingUsers, ", "))
	}
	players := []Player{}
	for _, player := range playersMap {
		players = append(players, player)
	}
	return players, nil
}

func processPlayers(players []Player) []Player {
	sort.Slice(players, func(i, j int) bool {
		return players[i].Diff < players[j].Diff
	})
	for i := 0; i < len(players); i++ {
		for j := len(players) - 1; j >= 0; j-- {
			payer := players[i]
			recipient := players[j]
			if recipient.Diff != 0 {
				if payer.Diff == 0 {
					break
				}
				if (payer.Diff + recipient.Diff) >= 0 {
					payer.Payments = append(payer.Payments, Payment{
						Amount:    math.Abs(payer.Diff),
						Recipient: &recipient,
					})
					recipient.Diff = recipient.Diff + payer.Diff
					payer.Diff = 0
				} else if (payer.Diff + recipient.Diff) < 0 {
					payer.Payments = append(payer.Payments, Payment{
						Amount:    math.Abs(recipient.Diff),
						Recipient: &recipient,
					})
					payer.Diff = recipient.Diff + payer.Diff
					recipient.Diff = 0
				}
			}
			players[i] = payer
			players[j] = recipient
		}
	}
	return players
}

func (p *Player) GetPaymentsComment() string {
	if len(p.Payments) == 0 {
		return ""
	}
	comment := ""
	for _, payment := range p.Payments {
		comment += fmt.Sprintf("%s -> %s %v руб на номер %s\n", p.Nickname, payment.Recipient.Nickname, payment.Amount, payment.Recipient.PaymentInfo)
	}
	comment += p.Login
	return comment
}

func getUserInfo(nickname string) *UserFromTable {
	for _, user := range users {
		cleanLogin := strings.Trim(user.Login, "@")
		if cleanLogin == nickname {
			return &user
		}
		if utils.Contains(user.Nicknames, nickname) {
			return &user
		}
	}
	return nil
}

func getComment(playersList []Player, date string) string {
	comment := "#table\nDate: " + date + "\n\n"
	items := []string{}
	for _, player := range playersList {
		if player.GetPaymentsComment() != "" {
			items = append(items, player.GetPaymentsComment())
		}
	}
	comment += strings.Join(items, "\n---------------\n")
	return comment
}
