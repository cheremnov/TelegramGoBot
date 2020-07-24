package main

import (
    "log"

    "github.com/go-telegram-bot-api/telegram-bot-api"

    "os"
    "encoding/json"
)

type BotSensitiveConfiguration struct{
    Token           string
}
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
type AuthorizedInfo struct{
    Users           []string
}
func getAuthorizedUsersJson( auth_info* AuthorizedInfo) error{
    const kAuthUserFile = "authorized.json"
    file, err := os.Open( kAuthUserFile)
    if err != nil{
        return err
    }
    defer file.Close()

    // Get current authorized users list
    decoder := json.NewDecoder( file)
    err = decoder.Decode( auth_info)
    if err != nil{
        return err
    }
    return nil
}
func addAuthorizedUser( user_name string) error{
    auth_info := AuthorizedInfo{}
    err_json := getAuthorizedUsersJson( &auth_info)
    if err_json != nil{
        return err_json
    }
    const kAuthUserFile = "authorized.json"
    file, err := os.OpenFile( kAuthUserFile, os.O_WRONLY | os.O_TRUNC, 0777)
    if err != nil{
        return err
    }
    defer file.Close()

    auth_info.Users = append( auth_info.Users, user_name)

    encoder := json.NewEncoder( file)
    err = encoder.Encode( &auth_info)
    if err != nil{
        return err
    }
    return nil
}
func isUserAuthorized( user_name string) bool{
    auth_info := AuthorizedInfo{}
    err_json := getAuthorizedUsersJson( &auth_info)
    if err_json != nil{
        log.Printf("%v", err_json)
        return false
    }
    for _, user := range auth_info.Users{
        if user_name == user{
            return true;
        }
    }
    return false
}
func main() {
    kKeyWord := "Credo in unum deum"
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
        /*if !isUserAuthorized( user_name){
            continue
        }*/

        log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
        msg := tgbotapi.NewMessage( update.Message.Chat.ID, update.Message.Text)
        msg.ReplyToMessageID = update.Message.MessageID

        if update.Message.Text == kKeyWord{
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
