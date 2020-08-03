package main

import (
    "log"
    "os"
    "encoding/json"
    "golang.org/x/crypto/bcrypt"
)
/**
 * (authorization.go) 
 * Contains the functions connected to the authorization.
 */
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
/**
 * Add a user to the authorized users list
 */
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
/**
 * Check the user authorization
 */
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
type KeywordInfo struct{
    AccessKey          string
}
const kKeywordFile = "keywords.json"
/**
 * Get a plain-text keyword.
 */
func getKeyword() string{
    keyword_info := KeywordInfo{}
    file, err := os.Open( kKeywordFile)
    if err != nil{
        log.Fatal( err)
    }
    defer file.Close()

    decoder := json.NewDecoder( file)
    err = decoder.Decode( &keyword_info)
    if err != nil{
        log.Fatal( err)
    }
    return keyword_info.AccessKey
}
/**
 * Generate a salted keyword
 */
func generateSaltedKeyword( orig_string string) (string, error){
    keyword := []byte( orig_string)
    salted_keyword, err := bcrypt.GenerateFromPassword( keyword, bcrypt.DefaultCost)
    if err != nil{
        return "", err
    }

    salted := string( salted_keyword[:])
    return salted, nil
}

type SaltedInfo struct{
    AccessKey           string
}
/**
 * Get a saved salted keyword
 */
func getSaltedKeyword() (string, error){
    file, err := os.OpenFile( kSaltedFile, os.O_RDONLY, 0777)
    if err != nil{
        return "", err
    }
    defer file.Close()
    salted_info := SaltedInfo{}
    decoder := json.NewDecoder( file)
    err = decoder.Decode( &salted_info)
    if err != nil{
        return "", err
    }
    return salted_info.AccessKey, nil
}
/**
 * Save a salted keyword to the file
 */
func setSaltedKeyword( salted_keyword string) error{
    file, err := os.OpenFile( kSaltedFile, os.O_WRONLY | os.O_TRUNC | os.O_CREATE, 0777)
    if err != nil{
        return err
    }
    defer file.Close()
    salted_info := SaltedInfo{}
    salted_info.AccessKey = salted_keyword
    encoder := json.NewEncoder( file)
    err = encoder.Encode( &salted_info)
    if err != nil{
        return err
    }
    return nil
}
/**
 * A wrapper on the bcrypt compare operation
 */
func checkSaltedKeyword( salted_keyword string, user_input string) error{
   /* salted_keyword, err := getSaltedKeyword()
    if err != nil{
        return err
    }*/
    salted_keyword1 := []byte( salted_keyword)
    user_input1 := []byte( user_input)
    return bcrypt.CompareHashAndPassword( salted_keyword1, user_input1)
}
