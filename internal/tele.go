package main

import (
	"carpool-btc/internal/app/utils"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotState int

var chatState = make(map[int64]BotState)

const (
	WaitingForNothing BotState = iota
	WaitingForEmail
	WaitingForPhoneNumber
	WaitingForDeposit
	WaitingForWalletAddress
)

type CoinGeckoResponse struct {
	Bitcoin struct {
		Usd float64 `json:"usd"`
	} `json:"bitcoin"`
}

type User struct {
	Email         string
	PhoneNumber   string
	Role          string
	UserName      string
	WalletAddress string
	RoleId        int64
	UserType      string
	Amount        int
}

var user User

func init() {
	utils.LoadEnvironmentVariable()
	utils.ConnectToDb()
}

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	// Loop through each update.
	for update := range updates {
		if update.Message != nil {
			handleMessage(update, bot)
		} else if update.CallbackQuery != nil {
			handleCallback(update, bot)
		}
	}

	user.updateDatabase()
}

func handleMessage(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	chatID := update.Message.Chat.ID
	user.UserName = update.Message.From.UserName

	if update.Message.IsCommand() {
		switch update.Message.Command() {
		case "register":
			msg := tgbotapi.NewMessage(chatID, "Please provide your email address: ")
			bot.Send(msg)
			chatState[chatID] = WaitingForEmail
		case "wallet":
			var walletAmount int
			var walletAddress string
			queryString := `SELECT wallet_address, amount FROM users WHERE email_address = ($1)`
			err := utils.DB.QueryRow(queryString, user.Email).Scan(&walletAddress, &walletAmount)
			if err != nil {
				log.Fatal(err.Error())
			}

			msg := tgbotapi.NewMessage(chatID, "Your Wallet.\nWallet Address: "+walletAddress+"\n"+"Balance: "+strconv.Itoa(walletAmount))
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
		}
	} else {
		switch chatState[chatID] {
		case WaitingForEmail:
			user.Email = update.Message.Text
			msg := tgbotapi.NewMessage(chatID, "Thank you, please now provide your phone number:")
			bot.Send(msg)
			chatState[chatID] = WaitingForPhoneNumber
		case WaitingForPhoneNumber:
			user.PhoneNumber = update.Message.Text
			msg := tgbotapi.NewMessage(chatID, "Thank you, please now provide your wallet address:")
			bot.Send(msg)
			chatState[chatID] = WaitingForWalletAddress
		case WaitingForWalletAddress:
			user.WalletAddress = update.Message.Text
			sendRoleSelection(bot, chatID)
			chatState[chatID] = WaitingForNothing
		case WaitingForDeposit:
			amount, err := strconv.Atoi(update.Message.Text)
			if err != nil {
				log.Fatal(err.Error())
			}

			user.Amount = amount
			user.updateWallet()
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "confirming transactions. Please wait")
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
			time.Sleep(5 * time.Second)
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, "successfully deposited. Please check the /wallet")
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
			chatState[chatID] = WaitingForNothing

		}

	}
}

func fetchBTCRate() (float64, error) {
	url := "https://api.coingecko.com/api/v3/simple/price?ids=bitcoin&vs_currencies=usd"
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var response CoinGeckoResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, err
	}

	return response.Bitcoin.Usd, nil
}

func usdToBTC(usdAmount float64) (float64, error) {
	rate, err := fetchBTCRate()
	if err != nil {
		return 0, err
	}
	return usdAmount / rate, nil
}

func handleCallback(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	query := update.CallbackQuery
	chatID := query.Message.Chat.ID

	user.UserType = query.Data
	switch query.Data {

	case "rider":
		sendWelcomeMessageWithButtons(bot, chatID)
		chatState[chatID] = WaitingForNothing
	case "deposit":
		chatState[chatID] = WaitingForNothing
		msg := tgbotapi.NewMessage(chatID, "Please specify the amount in USD: ")
		bot.Send(msg)
		chatState[chatID] = WaitingForDeposit
	case "driver":
		sendWelcomeMessageDriverWithButtons(bot, chatID)
		chatState[chatID] = WaitingForNothing
	}

	bot.Request(tgbotapi.CallbackConfig{
		CallbackQueryID: query.ID,
		Text:            "",
		ShowAlert:       false,
	})
}

func sendRoleSelection(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Are you a Driver or a Rider?")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Driver", "driver"),
			tgbotapi.NewInlineKeyboardButtonData("Rider", "rider"),
		),
	)

	bot.Send(msg)
}

func sendWelcomeMessageWithButtons(bot *tgbotapi.BotAPI, chatID int64) {
	user.updateDatabase()

	msg := tgbotapi.NewMessage(chatID, "Thank you for registering. Your email and phone number have been received. Please use the following command to interact with our bot")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Deposit", "deposit"),
			tgbotapi.NewInlineKeyboardButtonData("Withdraw", "withdraw"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("wallet", "profile"),
			tgbotapi.NewInlineKeyboardButtonData("Search Ride", "search"),
		),
	)
	bot.Send(msg)
}

func sendWelcomeMessageDriverWithButtons(bot *tgbotapi.BotAPI, chatID int64) {
	user.updateDatabase()

	msg := tgbotapi.NewMessage(chatID, "Thank you for registering. Your email and phone number have been received. Please use the following command to interact with our bot")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Balance", "balance"),
			tgbotapi.NewInlineKeyboardButtonData("Withdraw", "withdraw"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Create Ride", "create ride"),
			tgbotapi.NewInlineKeyboardButtonData("wallet", "profile"),
		),
	)
	bot.Send(msg)
}

func (u User) updateDatabase() {
	queryString := `
		INSERT INTO users (email_address, phone_number, user_name, wallet_address, user_type) 
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := utils.DB.Exec(queryString, u.Email, u.PhoneNumber, u.UserName, u.WalletAddress, u.UserType)

	if err != nil {
		log.Fatal("Failed to create database", err.Error())
	}
}

func (u User) updateWallet() {
	queryString := "UPDATE users SET amount = ($1) WHERE email_address = ($2)"

	_, err := utils.DB.Exec(queryString, u.Amount, u.Email)
	if err != nil {
		log.Fatal(err.Error())
	}
}
