package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/line/line-bot-sdk-go/v7/linebot"
)

type App struct {
	bot  *linebot.Client
	log  log.Logger
	addr string
}

func main() {
	// Create a line bot.
	bot, err := linebot.New(os.Getenv("CHANNEL_SECRET"), os.Getenv("CHANNEL_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	// Create an app.
	app := App{
		bot:  bot,
		log:  *log.Default(),
		addr: os.Getenv("SERVER_ADDR"),
	}

	// Start the server.
	app.log.Printf("Start server at %s\n", app.addr)
	http.HandleFunc("/callback", app.botHandler)
	app.log.Fatal(http.ListenAndServe(app.addr, nil))
}

func (app *App) botHandler(w http.ResponseWriter, r *http.Request) {
	events, err := app.bot.ParseRequest(r)
	if err != nil {
		switch {
		case errors.Is(err, linebot.ErrInvalidSignature):
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			var replyContent string

			// Check message type
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				replyContent = fmt.Sprintf(`You say "%s"`, message.Text)
			case *linebot.StickerMessage:
				replyContent = fmt.Sprintf(`Your sticker id is "%s", type is "%s"`, message.StickerID, message.StickerResourceType)
			default:
				replyContent = fmt.Sprintf(`Your message type is %s`, message.Type())
			}

			// Reply
			reply := linebot.NewTextMessage(replyContent)
			if _, err := app.bot.ReplyMessage(event.ReplyToken, reply).Do(); err != nil {
				app.log.Println(err)
			}
		}
	}
}
