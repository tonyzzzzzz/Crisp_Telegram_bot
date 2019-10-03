package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/crisp-im/go-crisp-api/crisp"
	"github.com/go-redis/redis"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/spf13/viper"
	"github.com/tonyzzzzzz/Crisp_Telegram_bot/utils"
)

var bot *tgbotapi.BotAPI
var client *crisp.Client
var redisClient *redis.Client
var config *viper.Viper

// CrispMessageInfo stores the original message
type CrispMessageInfo struct {
	WebsiteID string
	SessionID string
}

// MarshalBinary serializes CrispMessageInfo into binary
func (s *CrispMessageInfo) MarshalBinary() ([]byte, error) {
	return json.Marshal(s)
}

// UnmarshalBinary deserializes CrispMessageInfo into struct
func (s *CrispMessageInfo) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, s)
}

func contains(s []interface{}, e int64) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func replyToUser(update *tgbotapi.Update) {
	if update.Message.ReplyToMessage == nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "请回复一个消息")
		bot.Send(msg)
		return
	}

	res, err := redisClient.Get(strconv.Itoa(update.Message.ReplyToMessage.MessageID)).Result()

	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "ERROR: "+err.Error())
		bot.Send(msg)
		return
	}

	var msgInfo CrispMessageInfo
	err = json.Unmarshal([]byte(res), &msgInfo)

	if err := json.Unmarshal([]byte(res), &msgInfo); err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "ERROR: "+err.Error())
		bot.Send(msg)
		return
	}

	if update.Message.Text != "" {
		client.Website.SendTextMessageInConversation(msgInfo.WebsiteID, msgInfo.SessionID, crisp.ConversationTextMessageNew{
			Type:    "text",
			From:    "operator",
			Origin:  "chat",
			Content: update.Message.Text,
		})
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "回复成功！")
	bot.Send(msg)
}

func sendMsgToAdmins(text string, WebsiteID string, SessionID string) {
	for _, id := range config.Get("admins").([]interface{}) {
		msg := tgbotapi.NewMessage(id.(int64), text)
		msg.ParseMode = "Markdown"
		sent, _ := bot.Send(msg)

		redisClient.Set(strconv.Itoa(sent.MessageID), &CrispMessageInfo{
			WebsiteID,
			SessionID,
		}, 12*time.Hour)
	}
}

func init() {
	config = utils.GetConfig()

	log.Printf("Initializing Redis...")

	redisClient = redis.NewClient(&redis.Options{
		Addr:     config.GetString("redis.host"),
		Password: config.GetString("redis.password"),
		DB:       config.GetInt("redis.db"),
	})

	var err error

	_, err = redisClient.Ping().Result()
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Initializing Bot...")

	bot, err = tgbotapi.NewBotAPI(config.GetString("telegram.key"))

	if err != nil {
		log.Panic(err)
	}

	bot.Debug = config.GetBool("debug")
	bot.RemoveWebhook()

	log.Printf("Authorized on account %s", bot.Self.UserName)

	log.Printf("Initializing Crisp Listner")
	client = crisp.New()
	// Set authentication parameters
	client.Authenticate(config.GetString("crisp.identifier"), config.GetString("crisp.key"))

	// Connect to realtime events backend and listen (only to 'message:send' namespace)
	client.Events.Listen(
		[]string{
			"message:send",
		},

		func(reg *crisp.EventsRegister) {
			// Socket is connected: now listening for events

			// Notice: if the realtime socket breaks at any point, this function will be called again upon reconnect (to re-bind events)
			// Thus, ensure you only use this to register handlers

			// Register handler on 'message:send/text' namespace
			reg.On("message:send/text", func(evt crisp.EventsReceiveTextMessage) {
				text := fmt.Sprintf(`*%s(%s): *%s`, *evt.User.Nickname, *evt.User.UserID, *evt.Content)
				sendMsgToAdmins(text, *evt.WebsiteID, *evt.SessionID)
			})

			// Register handler on 'message:send/file' namespace
			reg.On("message:send/file", func(evt crisp.EventsReceiveFileMessage) {
				text := fmt.Sprintf(`*%s(%s): *[File](%s)`, *evt.User.Nickname, *evt.User.UserID, evt.Content.URL)
				sendMsgToAdmins(text, *evt.WebsiteID, *evt.SessionID)
			})

			// Register handler on 'message:send/animation' namespace
			reg.On("message:send/animation", func(evt crisp.EventsReceiveAnimationMessage) {
				text := fmt.Sprintf(`*%s(%s): *[Animation](%s)`, *evt.User.Nickname, *evt.User.UserID, evt.Content.URL)
				sendMsgToAdmins(text, *evt.WebsiteID, *evt.SessionID)
			})
		},

		func() {
			log.Printf("Crisp listener disconnected, reconnecting...")
		},

		func() {
			log.Fatal("Crisp listener error, check your API key or internet connection?")
		},
	)
}

func main() {
	var updates tgbotapi.UpdatesChannel

	log.Print("Start pooling")
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ = bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("%s %s: %s", update.Message.From.FirstName, update.Message.From.LastName, update.Message.Text)

		switch update.Message.Command() {
		case "start":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Blinkload Telegram 客服助手")
			msg.ParseMode = "Markdown"
			bot.Send(msg)
		}

		if contains(config.Get("admins").([]interface{}), int64(update.Message.From.ID)) {
			replyToUser(&update)
		}
	}
}
