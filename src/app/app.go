package main

import (
    "fmt"
    "net/http"
    "os"
    "time"
    "encoding/base64"
    "github.com/dgrijalva/jwt-go"
    "encoding/json"
    "github.com/gorilla/mux"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
)



// User : Коллекция пользователей для БД
type User struct{
    ID bson.ObjectId `bson:"_id"`
    Name string `bson:"name"`
    GUID string `bson:"guid"`
    Password string `bson:"password"`

}

// MyCustomClaims : поля для токена
type MyCustomClaims struct {
    exp int
    name string
	guid string
	jwt.StandardClaims
}

// Token : Коллекция токенов для БД
type Token struct {
    ID bson.ObjectId `bson:"_id"`
    TokensPair map [string]string `bson:"tokens"`
    GUID string `bson:"Guid_User"`
}

// TokensFromUser : токены для пользователя
type TokensFromUser struct {
    AccessToken string `json:"Access_Token"`
    RefreshToken string `json:"Refresh_Token"`

}


// Роутер
func main ()  {
    rout := mux.NewRouter()


    rout.Handle("/", http.FileServer(http.Dir("../utils/views/")))

    rout.HandleFunc("/css/{filename}", cssHandler)

    rout.HandleFunc("/js/{filename}", jsHandler)

    rout.HandleFunc("/get-tokens/{guid}", getTokensHandler).Methods("GET")

    rout.HandleFunc("/refresh-tokens/{guid}/{refresh-token}", refreshTokensHandler).Methods("GET")

    rout.HandleFunc("/delete-token/{guid}/{refresh-token}", deleteTokensHandler).Methods("GET")

    rout.HandleFunc("/delete-all-tokens/{guid}", deleteAllTokensHandler).Methods("GET")

    checkDB()


    port := os.Getenv("PORT")
        if port == "" {
            return
        }

    http.ListenAndServe(":"+port, rout)
}


//Контроллер подгрузки css
func cssHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    filename := vars["filename"]
    http.ServeFile(w, r, "../utils/css/"+filename)
}

//Контроллер подгрузки js
func jsHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    filename := vars["filename"]
    http.ServeFile(w, r, "../utils/js/"+filename)
}

// GetTokensHandler : контроллер создания пары токенов
func getTokensHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    guid := vars["guid"]
        //проверка пользователя
    currentUser := getOneUser(guid)
    if currentUser.ID == ""{
        w.Write([]byte ("Пользователь не найден"))
        fmt.Println("Пользователь не найден")
        return

    }
        //генерация пары токенов
    tokens := generateTokenPair(guid)
    accessToken := tokens["access_token"]
    refreshToken := tokens["refresh_token"]

        //хеширование рефреш токена и запись в базу пары токенов
    hashedRefreshToken := base64.StdEncoding.EncodeToString([]byte(refreshToken))
    tokens["refresh_token"] = hashedRefreshToken

    insertNewTokens(guid, tokens)


        //формирование ответа пользователю
    resp := TokensFromUser{accessToken, refreshToken}
    response, err := json.Marshal(resp)
    if err != nil {
        panic(err)
    }

    w.Write([]byte (response))

}


// refreshTokensHandler : контроллер обновления пары токенов
func refreshTokensHandler (w http.ResponseWriter, r *http.Request){
    vars := mux.Vars(r)
    guid := vars["guid"]
    refreshToken := vars["refresh-token"]

        //хеширование полученого токена

    hashedRefreshToken := base64.StdEncoding.EncodeToString([]byte(refreshToken))

        //получение токенов из БД и проверка, есть ли полученный токен в БД
    collectionTokens := getCollectionTokens(guid)

    var isTokenTrue = 0
    var idTokenInDB bson.ObjectId

    for i := 0; i < len(collectionTokens); i++{
        fmt.Println("collectionTokens:", collectionTokens[i].TokensPair["refresh_token"])

        if hashedRefreshToken == collectionTokens[i].TokensPair["refresh_token"]{
            idTokenInDB = collectionTokens[i].ID
            fmt.Println("idTokenInDB",idTokenInDB)
            isTokenTrue ++
        }
    }

    if isTokenTrue == 0{
        panic("Token is not good")
    }

            //проверка токена
        checkToken := checkRefreshToken(refreshToken)
        if checkToken == false{
            panic("Token is not good")
        }
            //создание новой пары токенов и замена в БД
        newTokenPair := generateTokenPair(guid)
        newAccessToken := newTokenPair["access_token"]
        newRefreshToken := newTokenPair["refresh_token"]


        hashedNewRefreshToken := base64.StdEncoding.EncodeToString([]byte(newTokenPair["refresh_token"]))
        newTokenPair["refresh_token"] = hashedNewRefreshToken
        fmt.Println("newTokenPair", newTokenPair)
        refreshTokensPair(guid, idTokenInDB, newTokenPair)


            //формирование ответа для пользователя
        resp := TokensFromUser{newAccessToken, newRefreshToken}
        response, err := json.Marshal(resp)
        if err != nil {
            panic(err)
        }

        w.Write([]byte (response))



//    newTokenPair := generateTokenPair(guid)



}

// deleteTokensHandler : контроллер удаления пары токенов не дороблено!!!!!!!
func deleteTokensHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    guid := vars["guid"]
    refreshToken := vars["refresh-token"]
        //проверка пользователя
    currentUser := getOneUser(guid)
    if currentUser.ID == ""{
        w.Write([]byte ("Пользователь не найден"))
        fmt.Println("Пользователь не найден")
        return
    }

        //хеширование полученого токена
    hashedRefreshToken := base64.StdEncoding.EncodeToString([]byte(refreshToken))

        //получение токенов из БД и проверка, есть ли полученный токен в БД
    collectionTokens := getCollectionTokens(guid)

    var isTokenTrue = 0
    var idTokenInDB bson.ObjectId

    for i := 0; i < len(collectionTokens); i++{
        fmt.Println("collectionTokens:", collectionTokens[i].TokensPair["refresh_token"])

        if hashedRefreshToken == collectionTokens[i].TokensPair["refresh_token"]{
            idTokenInDB = collectionTokens[i].ID
            fmt.Println("idTokenInDB",idTokenInDB)
            isTokenTrue ++
        }
    }
    if isTokenTrue == 0{
        panic("Token is not good")
    }

        //удаление текущего токена
    resultDelete := deleteTokens(guid, idTokenInDB)
    if resultDelete == false{
        w.Write([]byte ("Удаление не удалось"))
        fmt.Println("Удаление не удалось")
        return
    }
        //формирование ответа пользователю

    response, err := json.Marshal("Токен удален")
    if err != nil {
        panic(err)
    }
    w.Write([]byte (response))

}

// deleteAllTokensHandler : контроллер удаления всех токенов протестировать!!!!!!!
func deleteAllTokensHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    guid := vars["guid"]
        //проверка пользователя
    currentUser := getOneUser(guid)
    if currentUser.ID == ""{
        w.Write([]byte ("Пользователь не найден"))
        fmt.Println("Пользователь не найден")
        return

    }

        //удаление всех токенов для этого пользователя

    resultDelete := deleteTokens(guid, "")
    if resultDelete == false{
        w.Write([]byte ("Удаление не удалось"))
        fmt.Println("Удаление не удалось")
        return
    }

        //формирование ответа пользователю
    response, err := json.Marshal("Токены удалены")
    if err != nil {
        panic(err)
    }

    w.Write([]byte (response))

}





//Функция создания пары токенов
func generateTokenPair(guid string) (map[string]string) {


        accessToken := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
            "exp": time.Now().Add(time.Minute * 24).Unix(),
            "name": "User1",
            "guid": guid,
        })

        accessTokenString, err := accessToken.SignedString([]byte("secret"))
        fmt.Println(err)





        refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
        //    "exp": time.Now().Add(time.Second * 24 ).Unix(),
            "exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
            "guid": guid,
        })

        refreshTokenString, err := refreshToken.SignedString([]byte("secret"))
        if err != nil {
            panic(err)
        }

        return  map [string] string{
            "access_token":  accessTokenString,
            "refresh_token": refreshTokenString,
        }
}
//Функция проверки refresh токена
func checkRefreshToken(refreshToken string)(tokenIsValid bool){
    refreshTokenParse, err := jwt.Parse(refreshToken, func(refreshTokenParse *jwt.Token) (interface{}, error) {


    if _, ok := refreshTokenParse.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, fmt.Errorf("Unexpected signing method: %v", refreshTokenParse.Header["alg"])
    }

        return ([]byte("secret")), nil
    })
    if err != nil {
        panic(err)
    }


    if claims, ok := refreshTokenParse.Claims.(jwt.MapClaims); ok && refreshTokenParse.Valid {
        fmt.Println(claims)
        tokenIsValid = true
        return tokenIsValid
    }
    tokenIsValid = false
    return tokenIsValid

}




    //Работа с БД

    //проверка БД на наличие таблицы пользователей
func checkDB()  {
    //session, err := mgo.Dial("mongodb://user1:W7XMgiTE#c6_KmL@ds141068.mlab.com:41068/heroku_1hbzfgwk")
    //session, err := mgo.Dial("mongodb://user1:W7XMgiTE#c6_KmL@ds141068-a0.mlab.com:41068, ds141068-a1.mlab.com:41069/heroku_1hbzfgwk?replicaSet=rs-ds141068")
    session, err := mgo.Dial("mongodb://heroku_0g8rdn33:vtf590ps7pv7e9q1ffk56pq2h6@ds243502.mlab.com:43502/heroku_0g8rdn33")
    if err != nil {
        panic(err)
    }
    defer session.Close()
    userCollection := session.DB ("heroku_1hbzfgwk").C("users")

    query := bson.M {}

    users := []User{}

    userCollection.Find(query).All(&users)

    if len(users) == 0{
        user1 := &User{ID:bson.NewObjectId(), Name:"UserOne", GUID:"dd7f8228-2f45-4b63-9bdc-87989e693204", Password:"96E79218965EB72C92A549DD5A330112"}

        err = userCollection.Insert(user1)
        if err != nil{
            fmt.Println(err)
        }
    } else {
        for _, u := range users{

            fmt.Println(u.Name, u.GUID, "БД уже была создана")
        }
    }

}
    //получение токенов
func getCollectionTokens(guid string)( []Token){

    session, err := mgo.Dial("mongodb://user1:W7XMgiTE#c6_KmL@ds141068.mlab.com:41068/heroku_1hbzfgwk")
    if err != nil {
        panic(err)
    }
    defer session.Close()
    tokenCollection := session.DB ("heroku_1hbzfgwk").C("tokens")
    query := bson.M{"Guid_User": guid}
    tokensCurrentUser := []Token{}
    tokenCollection.Find(query).All(&tokensCurrentUser)
    return tokensCurrentUser
}
    //получение пользователя
func getOneUser(guid string) (User) {
    session, err := mgo.Dial("mongodb://user1:W7XMgiTE#c6_KmL@ds141068.mlab.com:41068/heroku_1hbzfgwk")
    if err != nil {
        panic(err)
    }
    defer session.Close()
    userCollection := session.DB ("heroku_1hbzfgwk").C("users")
    query := bson.M {"guid": guid}
    var user = User{}
    userCollection.Find(query).One(&user)
    return user
}
    //добавление пары токенов
func insertNewTokens(guid string, tokens map[string]string)  {

    session, err := mgo.Dial("mongodb://user1:W7XMgiTE#c6_KmL@ds141068.mlab.com:41068/heroku_1hbzfgwk")
    if err != nil {
        panic(err)
    }
    defer session.Close()
    tokenCollection := session.DB ("heroku_1hbzfgwk").C("tokens")

    token1 := &Token{ID:bson.NewObjectId(), TokensPair:tokens, GUID: guid}

    err = tokenCollection.Insert(token1)
    if err != nil{
        fmt.Println(err)
    }
}
    //обновление пары токенов
func refreshTokensPair(guid string, idTokenInDB bson.ObjectId, newTokenPair map[string]string)  {

    session, err := mgo.Dial("mongodb://user1:W7XMgiTE#c6_KmL@ds141068.mlab.com:41068/heroku_1hbzfgwk")
    if err != nil {
        panic(err)
    }
    defer session.Close()
    tokenCollection := session.DB ("heroku_1hbzfgwk").C("tokens")



    query := bson.M {"Guid_User": guid, "_id": idTokenInDB}


    tokenCollection.Update(query, bson.M{"$set":bson.M{"tokens": newTokenPair}})




}
    //удаление токенов
func deleteTokens(guid string, idTokenInDB bson.ObjectId) (bool)  {

    session, err := mgo.Dial("mongodb://user1:W7XMgiTE#c6_KmL@ds141068.mlab.com:41068/heroku_1hbzfgwk")
    if err != nil {
        panic(err)
    }
    defer session.Close()
    tokenCollection := session.DB ("heroku_1hbzfgwk").C("tokens")



    if idTokenInDB == ""{
        query := bson.M{"Guid_User": guid}
        _, err = tokenCollection.RemoveAll(query)
        if err != nil{
            fmt.Println(err)
            return false
        }
        return true
    }
        query := bson.M {"Guid_User": guid, "_id": idTokenInDB}
        _, err = tokenCollection.RemoveAll(query)
        if err != nil{
            fmt.Println(err)
            return false
        }
        return true
}
