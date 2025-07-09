package main

import (
	"asyasocute/anonimous-questions/config"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	ID       int64 `gorm:"primaryKey"`
	Receiver int64
}

type Link struct {
	Text   string `gorm:"primaryKey"`
	UserID int64
}

func getUserLink(UserID int64, db *gorm.DB) string {
	var link Link
	db.First(&link, "user_id = ?", UserID)
	if link.UserID != UserID {
		link = Link{Text: strconv.FormatInt(UserID+1524181, 16), UserID: UserID}
		db.Create(&link)
	}
	return link.Text
}
func getUserFromLink(key string, db *gorm.DB) (int64, error) {
	var link Link
	db.First(&link, "text = ?", key)
	if link.Text == key {
		fmt.Println(link.UserID)
		return link.UserID, nil
	}
	return 0, errors.New("no user with this link")
}

func main() {
	config.Load()

	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&User{})
	db.AutoMigrate(&Link{})

	bot, err := telego.NewBot(config.C.BotApiToken, telego.WithDefaultLogger(true, true))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	updates, _ := bot.UpdatesViaLongPolling(ctx, nil)

	bh, _ := th.NewBotHandler(bot, updates)

	bh.HandleMessage(func(ctx *th.Context, msg telego.Message) error {
		start := strings.Split(msg.Text, " ")
		if len(start) == 2 {
			userId, err := getUserFromLink(start[1], db)
			if userId == msg.Chat.ID {
				bot.SendMessage(ctx, tu.Messagef(
					tu.ID(msg.Chat.ID),
					"You can't text yourself",
				))
				return nil
			} else if err != nil {
				bot.SendMessage(ctx, tu.Messagef(
					tu.ID(msg.Chat.ID),
					"Something went wrong...",
				))
				return nil
			}
			db.Delete(&User{}, msg.From.ID)
			db.Create(&User{ID: msg.From.ID, Receiver: userId})
			bot.SendMessage(ctx, tu.Messagef(
				tu.ID(msg.Chat.ID),
				"Write your message to user!",
			))

		} else {
			userLink := getUserLink(msg.From.ID, db)
			bot.SendMessage(ctx, tu.Messagef(
				tu.ID(msg.Chat.ID),
				"Hi, your link is: https://t.me/%s?start=%s", config.C.BotUsername, userLink,
			))
		}
		return nil
	}, th.CommandEqual("start"))
	bh.HandleMessage(func(ctx *th.Context, msg telego.Message) error {
		var receiver User
		db.Take(&receiver, msg.Chat.ID)
		fmt.Println(receiver)
		bot.SendMessage(ctx, tu.Messagef(
			tu.ID(receiver.Receiver),
			"New message!\n\n%s", msg.Text,
		))
		bot.SendMessage(ctx, tu.Messagef(
			tu.ID(msg.From.ID),
			"Delivered. You can write another one",
		))
		return nil
	})

	fmt.Println("Started")
	defer func() { _ = bh.Stop() }()
	_ = bh.Start()
	fmt.Println("Done")
}
