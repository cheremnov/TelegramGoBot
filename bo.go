package main

import (
    "log"

    "github.com/go-telegram-bot-api/telegram-bot-api"
    "os"
    "encoding/json"

    "time"
    "strings"
    "bufio"
    "errors"

    "github.com/cheremnov/TelegramGoBot/model"
)

// File with a salted keyword
const kSaltedFile = "salted.json"

// Update notifications every kNotificationsRefreshTime seconds
const kNotificationsRefreshTime = 10
// An UTC timezone
const kUTCTimeZone = 3

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
/**
 * Send all expired notifications
 * Results:
       error, if notifications aren't sent
 */
func notificationsSend( bot* tgbotapi.BotAPI, model_ctx* model.ModelContext) error{
    /**
     * For one iteration of sending the notifications
     * the expiration time must remain constant
     */
    expiration_time := time.Now()
    // Get expired notifications
    rows, err := model.GetExpiredNotifications( expiration_time, model_ctx)
    if err != nil {
        log.Panic( err)
    }
    for rows.Next(){
        message_user_id := 0
        message := ""
        if err := rows.Scan( &message_user_id, &message); err != nil{
            log.Panic( err)
            return err
        }
        chat_id, err := model.GetChatId( message_user_id, model_ctx)
        if err == nil{
            msg := tgbotapi.NewMessage( chat_id, message)
            bot.Send( msg)
        }
        log.Printf( "User id: %v Notification: %v\n", message_user_id, message)
    }
    err = model.DeleteExpiredNotifications( expiration_time, model_ctx)
    if err != nil {
        log.Panic( err)
    }
    return nil
}
/**
 * Create a goroutine that sends the notifications every 10 seconds
 * TODO: Check the concurrency issues
 */
func createNotificationGoRoutine( bot* tgbotapi.BotAPI, model_ctx* model.ModelContext){
    // Set a notifications goroutine
    ticker := time.NewTicker( kNotificationsRefreshTime * time.Second)
    /**
     * Send a message to quit channel to stop the goroutine
     * Currently unused
     */
    quit := make(chan struct{})
    go func() {
        for{
            select{
            case <- ticker.C:
                notificationsSend( bot, model_ctx)
            case <- quit:
                ticker.Stop()
                return
            }
        }
    }()
}
/**
 * Every notification message string should have this format:
 *  "/notification [NOTIFICATION_TIME] [MESSAGE]
 * Notification time is "YYYY-MM-DD-HH:MM:SS"
 * Results:
 *     A notification time, message and an error.
 */
func parseNotification( notification_command string) (time.Time, string, error) {
    notification_slice := strings.Split( notification_command, "")
    message := ""
    notification_time := time.Time{}
    if len(notification_slice) < 2 {
        return notification_time, message, errors.New("Wrong notification string format")
    }
    // Use a bufio library to parse the string
    scanner := bufio.NewScanner( strings.NewReader( notification_command))
    parsed_time := false
    var err error
    // Split by the words
    scanner.Split( bufio.ScanWords)
    /**
     * Look for the time in the specified layout form.
     * When found, start recording the message
     */
    for scanner.Scan() {
        if !parsed_time {
            // Layout for "YYYY-MM-DD-HH:MM:SS"
            layout := "2006-01-02-15:04:05"
            notification_time, err = time.Parse( layout, scanner.Text() )
            if err == nil {
                parsed_time = true
            }
        } else {
            message += scanner.Text() + " "
        }
    }
    if err := scanner.Err(); err != nil || !parsed_time {
        return notification_time, message, errors.New( "Wrong notification string format")
    }
    // Account for the timezones
    return notification_time.Add( -kUTCTimeZone * time.Hour), message, nil
}
func main() {
    model_ctx, err := model.ModelInit()
    if err != nil{
        log.Panic( err)
    }
    defer model.ModelTerminate( &model_ctx)

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

    /**
     * The bot uses two goroutines: 
     * for receiving updates and for sending notifications
     */
    createNotificationGoRoutine( bot, &model_ctx)

    updates, err := bot.GetUpdatesChan(u)

    for update := range updates {
        if update.Message == nil { // ignore any non-Message Updates
            continue
        }
        user_name := update.Message.From.UserName
        user_id := update.Message.From.ID
        chat_id := update.Message.Chat.ID
        // Save a private chat for user
        if update.Message.Chat.IsPrivate(){
            err = model.SavePrivateChat( user_id, chat_id, &model_ctx)
            if err != nil{
                log.Panic( err)
            }
        }


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
        } else{
            // A user authorized
            if update.Message.IsCommand(){
                // Only notification command is available now
                if update.Message.Command() == "notification" {
                    notification_time, notification_message, err := parseNotification( update.Message.Text)
                    if err != nil {
                        msg = tgbotapi.NewMessage( update.Message.Chat.ID, "The wasp can't decipher your notification")
                    } else{
                        msg = tgbotapi.NewMessage( update.Message.Chat.ID, "The wasp will notify you")
                        model.SaveNotification( notification_message, notification_time, user_id, &model_ctx)
                    }
                }
            }
        }

        bot.Send(msg)

    }
}
