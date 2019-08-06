package main

import (
	"context"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	log.Printf("Starting Chat4Bread Backend.")

	// Connect with MongoDB
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	dburi := fmt.Sprintf("mongodb://%s:%s@database:27017", os.Getenv("MONGO_USERNAME"), os.Getenv("MONGO_PASSWORD"))
	log.Printf("Connecting to MongoDB database: %s.", dburi)
	db, err := mongo.Connect(ctx, options.Client().ApplyURI(dburi))
	if err != nil {
		log.Panic(err)
	}
	err = db.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Connected to MongoDB database.")

	// Setup state machine
	orm := NewORM(db, "chat4bread")
	err = orm.CreateIndicies()
	if err != nil {
		log.Panic(err)
	}

	cai := NewCAI(os.Getenv("CAI_TOKEN"))
	machine := NewMachine(orm, cai)

	// Connect with Telegram
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	if err != nil {
		log.Panic(err)
	}
	//bot.Debug = true
	machine.SendMessage = func(id int64, message string) error {
		msg := tgbotapi.NewMessage(id, message)
		_, err := bot.Send(msg)
		return err
	}

	// Process messages
	log.Printf("Authorized on Telegram bot account %s", bot.Self.UserName)

	var updates tgbotapi.UpdatesChannel

	if os.Getenv("TELEGRAM_WEBHOOK_URL") == "" {
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60
		updates, err = bot.GetUpdatesChan(u)
		if err != nil {
			log.Panic(err)
		}
	} else {
		_, err = bot.SetWebhook(tgbotapi.NewWebhook(os.Getenv("TELEGRAM_WEBHOOK_URL") + bot.Token))
		if err != nil {
			log.Fatal(err)
		}
		info, err := bot.GetWebhookInfo()
		if err != nil {
			log.Fatal(err)
		}
		if info.LastErrorDate != 0 {
			log.Printf("Telegram callback last failed: %s", info.LastErrorMessage)
		}
		updates = bot.ListenForWebhook("/" + bot.Token)
		go http.ListenAndServe("0.0.0.0:8080", nil)
	}
	for update := range updates {
		//log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		reply, err := machine.Generate(update.Message.Chat.ID, update.Message.Text)
		if err != nil {
			log.Printf("Error: %s", err.Error())
			reply = fmt.Sprintf("Error: %s", err.Error())
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
		_, err = bot.Send(msg)
		if err != nil {
			log.Panic(err)
		}
	}

	log.Printf("Stopping Chat4Bread Backend.")
}
