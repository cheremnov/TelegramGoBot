package model

import (
    "database/sql"
    "time"

    "os"
    "encoding/json"

   // "fmt"
)

import _"github.com/go-sql-driver/mysql"

/**
 * The model module. Connects to the database
 * Works with the data on a server part. 
 */
type DatabaseConfiguration struct{
    Password           string
    ProtocolType       string
    Server             string
    DatabaseName       string
}
/**
 * A model context. Stores information about the model
 * Every function in the module depends on it.
 */
type ModelContext struct{
    database_          *sql.DB
    configuration_     DatabaseConfiguration
}

/**
 * Get a SQL database configuration
 */
func getDatabaseConfig( configuration* DatabaseConfiguration) error{
    const kDatabaseConfig = "model/databaseConfig.json"
    file, err := os.Open( kDatabaseConfig)
    if err != nil{
        return err
    }
    // On the return close the file
    defer file.Close()
    decoder := json.NewDecoder( file)
    err = decoder.Decode( configuration)
    if err != nil{
        return err
    }
    return nil
}
/**
 * Initialize a model
 */
func ModelInit() (ModelContext, error){
    model_ctx := ModelContext{}
    configuration := DatabaseConfiguration{}
    err := getDatabaseConfig( &configuration)
    if err != nil {
        return model_ctx, err
    }
    // Construct a DSN string to connect for a server
    dsn_string := ("TelegramBot:" + configuration.Password + "@" + configuration.ProtocolType) +
        "(" + configuration.Server + ")/" + configuration.DatabaseName
    db, err := sql.Open("mysql", dsn_string);
    if err != nil {
        return model_ctx, err
    }
    model_ctx.database_ = db
    model_ctx.configuration_ = configuration
    return model_ctx, nil
}
/**
 * Free all resources, allocated to the model
 */
func ModelTerminate( model_ctx* ModelContext) error{
    db := model_ctx.database_
    db.Close()
    return nil
}
/**
 * Save a notification to the SQL database
 * Results:
 *      error, if not saved
 */
func SaveNotification( message string, time_notification time.Time, user_id int, model_ctx* ModelContext) error{
    // Omit the explicit dereferencing of a pointer
    db := model_ctx.database_
    configuration := model_ctx.configuration_
    _, err := db.Exec("INSERT INTO " + configuration.DatabaseName + ".Notifications(user_id, message, notification_time) VALUES(?, ?, ?); ", user_id, message, time_notification)
    if err != nil {
        return err
    }
    return nil
}

/**
 * Get all expired notifications
 * Notification expires, when its time comes
 * Results:
       A list of the expired notifications
 */
func GetExpiredNotifications( expiration_time time.Time, model_ctx* ModelContext) ( *sql.Rows, error){
    db := model_ctx.database_
    configuration := model_ctx.configuration_
    rows, err := db.Query( " SELECT user_id, message FROM " + configuration.DatabaseName +
        ".Notifications WHERE notification_time < ?", expiration_time)
    if err != nil{
        return nil, err
    }
    return rows, nil
}
/**
 * Delete all expired notifications
 * Results:
 *     An error, if not deleted correctly
 */
func DeleteExpiredNotifications( expiration_time time.Time, model_ctx* ModelContext) error {
    db := model_ctx.database_
    configuration := model_ctx.configuration_
    _, err := db.Exec( "DELETE FROM " + configuration.DatabaseName +
        ".Notifications WHERE notification_time < ?", expiration_time)
    if err != nil{
        return err
    }
    return nil
}

/**
 * It's possible to send a message only to a chat, not to a user.
 * Keep in the database the user's private chat.
 * Assume no user has several chats. Assume a chat id is fixed.
 * TODO: Check if the constraints are valid
 */
/**
 * Save a chat if necessary
 */
func SavePrivateChat( user_id int, chat_id int64, model_ctx* ModelContext) error {
    db := model_ctx.database_
    configuration := model_ctx.configuration_
    var is_chat_saved bool
    // Does the user have a saved private chat
    err := db.QueryRow( "SELECT IF( COUNT(*), 'true', 'false') FROM " + configuration.DatabaseName +
        ".UserPrivateChats WHERE user_id = ?", user_id).Scan( &is_chat_saved)
    if err != nil{
        return err
    }
    // Save the chat
    if is_chat_saved == false {
        _, err := db.Exec( "INSERT INTO " + configuration.DatabaseName + ".UserPrivateChats( user_id, chat_id) VALUES (?, ?)", user_id, chat_id)
        if err != nil{
            return err
        }
    }
    return nil
}
/**
 * Get a chat id from a user id
 */
func GetChatId( user_id int, model_ctx* ModelContext) (int64, error) {
    db := model_ctx.database_
    configuration := model_ctx.configuration_
    var chat_id int64
    err := db.QueryRow( "SELECT chat_id FROM " + configuration.DatabaseName + ".UserPrivateChats WHERE user_id = ?", user_id).Scan( &chat_id)
    if err != nil{
        return 0, err
    }
    return chat_id, nil
}
