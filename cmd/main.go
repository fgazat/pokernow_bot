package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fgazat/poker/pkg/calc"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	bot           *tgbotapi.BotAPI
	usersJsonPath string
)

// Menu texts
// firstMenu  = "<b>Menu 1</b>\n\nA beautiful menu with a shiny inline button."
// secondMenu = "<b>Menu 2</b>\n\nA better menu with even more shiny inline buttons."
//
// // Button texts
// nextButton     = "Next"
// backButton     = "Back"
// tutorialButton = "Tutorial"

// Store bot screaming status

// Keyboard layout for the first menu. One button, one row
// firstMenuMarkup = tgbotapi.NewInlineKeyboardMarkup(
// 	tgbotapi.NewInlineKeyboardRow(
// 		tgbotapi.NewInlineKeyboardButtonData(nextButton, nextButton),
// 	),
// )
// // Keyboard layout for the second menu. Two buttons, one per row
// secondMenuMarkup = tgbotapi.NewInlineKeyboardMarkup(
// 	tgbotapi.NewInlineKeyboardRow(
// 		tgbotapi.NewInlineKeyboardButtonData(backButton, backButton),
// 	),
// 	tgbotapi.NewInlineKeyboardRow(
// 		tgbotapi.NewInlineKeyboardButtonURL(tutorialButton, "https://core.telegram.org/bots/api"),
// 	),
// )

func main() {
	var err error
	token := os.Getenv("TG_TOKEN")
	bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}
	usersJsonPath = os.Getenv("USERS_JSON_PATH")
	if usersJsonPath == "" {
		log.Panic("please specify USERS_JSON_PATH")
	}

	// Set this to true to log all interactions with telegram servers
	bot.Debug = true

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// Create a new cancellable background context. Calling `cancel()` leads to the cancellation of the context
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	// `updates` is a golang channel which receives telegram updates
	updates := bot.GetUpdatesChan(u)

	// Pass cancellable context to goroutine
	go receiveUpdates(ctx, updates)

	// Tell the user the bot is online
	log.Println("Start listening for updates. Press enter to stop")

	// Wait for a newline symbol, then cancel handling updates
	bufio.NewReader(os.Stdin).ReadBytes('\n')
	cancel()
}

func receiveUpdates(ctx context.Context, updates tgbotapi.UpdatesChannel) {
	// `for {` means the loop is infinite until we manually stop it
	for {
		select {
		// stop looping if ctx is cancelled
		case <-ctx.Done():
			return
		// receive update from channel and then handle it
		case update := <-updates:
			handleUpdate(update)
		}
	}
}

func handleUpdate(update tgbotapi.Update) {
	switch {
	// Handle messages
	case update.Message != nil:
		handleMessage(update.Message)
		// Handle button clicks
		// case update.CallbackQuery != nil:
		// 	handleButton(update.CallbackQuery)
	}
}

func handleMessage(message *tgbotapi.Message) {
	user := message.From
	text := message.Text

	if user == nil {
		return
	}
	// Print to console
	log.Printf("%s wrote %s", user.FirstName, text)

	var err error
	if strings.HasPrefix(text, "/") {
		err = handleCommand(message.Chat.ID, text)
	}
	// } else if screaming && len(text) > 0 {
	// 	// msg := tgbotapi.NewMessage(message.Chat.ID, strings.ToUpper(text))
	// 	// To preserve markdown, we attach entities (bold, italic..)
	// 	msg.Entities = message.Entities
	// 	_, err = bot.Send(msg)
	// }
	//    else {
	// 	// This is equivalent to forwarding, without the sender's name
	// 	copyMsg := tgbotapi.NewCopyMessage(message.Chat.ID, message.Chat.ID, message.MessageID)
	// 	_, err = bot.CopyMessage(copyMsg)
	// }

	if err != nil {
		log.Printf("An error occured: %s", err.Error())
	}
}

func reportErr(chatID int64, err error) {
	msg := tgbotapi.NewMessage(chatID, err.Error())
	msg.ParseMode = tgbotapi.ModeMarkdown
	bot.Send(msg)
}

// When we get a command, we react accordingly
func handleCommand(chatId int64, text string) error {
	parts := strings.Split(text, " ")
	command := parts[0]
	switch command {
	case "/calc":
		if len(parts) == 1 {
			err := fmt.Errorf("Please specify url `/calc URL`")
			reportErr(chatId, err)
			return err
		}
		comment, err := calc.Calcuclate(parts[1], usersJsonPath, calc.DownloadImpl{})
		if err != nil {
			reportErr(chatId, err)
			return err
		}
		msg := tgbotapi.NewMessage(chatId, comment)
		_, err = bot.Send(msg)
		if err != nil {
			return err
		}

	case "/new":
		if len(parts) < 4 {
			err := fmt.Errorf("some of the parameters are missing. Example of command: /new IN_GAME_NICKNAME LOGIN PAYMENT_INFO")
			reportErr(chatId, err)
			return err
		}
		err := calc.New(parts[1], parts[2], parts[3], usersJsonPath)
		if err != nil {
			reportErr(chatId, err)
			return err
		}
		msg := tgbotapi.NewMessage(chatId, "Successfully created new user.")
		_, err = bot.Send(msg)
		if err != nil {
			return err
		}

	case "/map":
		if len(parts) < 3 {
			err := fmt.Errorf("some of the parameters are missing. Example of command: /new IN_GAME_NICKNAME LOGIN PAYMENT_INFO")
			reportErr(chatId, err)
			return err
		}
		err := calc.Map(parts[1], parts[2], usersJsonPath)
		if err != nil {
			reportErr(chatId, err)
			return err
		}
		msg := tgbotapi.NewMessage(chatId, "Successfully created new user.")
		_, err = bot.Send(msg)
		if err != nil {
			return err
		}
	case "/help":
		return SendHelp(chatId)
		// case "/menu":
		// 	err = sendMenu(chatId)
		// 	break
	}
	return nil
}

// func handleButton(query *tgbotapi.CallbackQuery) {
// 	var text string
//
// 	markup := tgbotapi.NewInlineKeyboardMarkup()
// 	message := query.Message
//
// 	// if query.Data == nextButton {
// 	// 	text = secondMenu
// 	// 	markup = secondMenuMarkup
// 	// } else if query.Data == backButton {
// 	// 	text = firstMenu
// 	// 	markup = firstMenuMarkup
// 	// }
//
// 	callbackCfg := tgbotapi.NewCallback(query.ID, "")
// 	bot.Send(callbackCfg)
//
// 	// Replace menu text and keyboard
// 	msg := tgbotapi.NewEditMessageTextAndMarkup(message.Chat.ID, message.MessageID, text, markup)
// 	msg.ParseMode = tgbotapi.ModeHTML
// 	bot.Send(msg)
// }

// func sendMenu(chatId int64) error {
// 	// msg := tgbotapi.NewMessage(chatId, firstMenu)
// 	// msg.ParseMode = tgbotapi.ModeHTML
// 	// msg.ReplyMarkup = firstMenuMarkup
// 	// _, err := bot.Send(msg)
// 	return err
// }

func SendHelp(chatId int64) error {
	msg := tgbotapi.NewMessage(chatId, "Please check the README: https://github\\.com/fgazat/pokernow")
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	_, err := bot.Send(msg)
	return err
}
