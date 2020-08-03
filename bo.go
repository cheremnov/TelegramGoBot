package main

import (
    "log"

    "github.com/go-telegram-bot-api/telegram-bot-api"

    "os"
    "encoding/json"
)

// File with a salted keyword
const kSaltedFile = "salted.json"

type BotSensitiveConfiguration struct{
    Token           string
}
/**
 * Get the token from the config file.
 */
func getToken() string{
    kConfigAddress := "config.json"
    file, err := os.Open( kConfigAddress)
    if err != nil{
        log.Fatal(err)
    }
    // On the return close the file
    defer file.Close()
    decoder := json.NewDecoder( file)
    configuration := BotSensitiveConfiguration{}
    err2 := decoder.Decode( &configuration)
    if err2 != nil{
        log.Fatal( err2)
    }
    return configuration.Token
}
func main() {
    // Either file with a plain-text or a salted keyword must exist
    // TODO: Check it properly
    salted_keyword := ""
    if _, err := os.Stat( kSaltedFile); os.IsNotExist( err){
        keyword := getKeyword()
        salted_keyword, err = generateSaltedKeyword( keyword)
        if err != nil {
            log.Panic( err)
        }
        err = setSaltedKeyword( salted_keyword)
        if err != nil {
            log.Panic( err)
        }
    } else{
        salted_keyword, err = getSaltedKeyword()
        if err != nil {
            log.Panic( err)
        }
    }

    token := getToken()
    bot, err := tgbotapi.NewBotAPI( token)
    if err != nil {
        log.Panic(err)
    }

    bot.Debug = true

    log.Printf("Authorized on account %s", bot.Self.UserName)

    u := tgbotapi.NewUpdate(0)
    u.Timeout = 60

    updates, err := bot.GetUpdatesChan(u)

    for update := range updates {
        if update.Message == nil { // ignore any non-Message Updates
            continue
        }
        user_name := update.Message.From.UserName

        log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
        msg := tgbotapi.NewMessage( update.Message.Chat.ID, update.Message.Text)
        msg.ReplyToMessageID = update.Message.MessageID

        if checkSaltedKeyword( salted_keyword, update.Message.Text) == nil{
           if !isUserAuthorized( user_name){
                err = addAuthorizedUser( user_name)
                if err != nil{
                    log.Printf( "%v", err)
                }
                msg = tgbotapi.NewMessage( update.Message.Chat.ID, "The wasp heeds your call")
            } else{
                msg = tgbotapi.NewMessage( update.Message.Chat.ID, "The wasp always heeds your call")
            }
        } else if !isUserAuthorized( user_name){
            msg = tgbotapi.NewMessage( update.Message.Chat.ID, "The wasp won't answer you")
        }

        bot.Send(msg)

    }
}
