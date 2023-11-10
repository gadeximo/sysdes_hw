package service
 
import (
    "net/http"
	"crypto/sha256"
    "github.com/gin-gonic/gin"
	database "todolist.go/db"
    "strings"
)
 
func NewUserForm(ctx *gin.Context) {
    ctx.HTML(http.StatusOK, "new_user_form.html", gin.H{"Title": "Register user"})
}

func hash(pw string) []byte {
    const salt = "todolist.go#"
    h := sha256.New()
    h.Write([]byte(salt))
    h.Write([]byte(pw))
    return h.Sum(nil)
}

func RegisterUser(ctx *gin.Context) {
    // フォームデータの受け取り
    username := ctx.PostForm("username")
    password := ctx.PostForm("password")
	passwordRe:= ctx.PostForm("passwordRe")
    switch {
    case username == "":
        ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Usernane is not provided", "Password": password})
		return
    case password == "":
        ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password is not provided", "Username": username})
		return
	case password != passwordRe:
        ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "confirm password did not match ", "Username": username, "Password": password})
		return
    case !isPasswordComplex(password):
        ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password is not complex, please use at least one letter (A-Z a-z) and least one digit (0-9) and special character (!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~)", "Username": username})
        return
    case len(password)<10:
        ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password is short, please at least 10 characters", "Username": username})
        return
    }
    
    // DB 接続
    db, err := database.GetConnection()
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
	// 重複チェック
    var duplicate int
    err = db.Get(&duplicate, "SELECT COUNT(*) FROM users WHERE name=?", username)
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    if duplicate > 0 {
        ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Username is already taken", "Username": username, "Password": password})
        return
    }
 
    // DB への保存
    result, err := db.Exec("INSERT INTO users(name, password) VALUES (?, ?)", username, hash(password))
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
 
    // 保存状態の確認
    id, _ := result.LastInsertId()
    var user database.User
    err = db.Get(&user, "SELECT id, name, password FROM users WHERE id = ?", id)
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    ctx.JSON(http.StatusOK, user)
}

func isPasswordComplex(password string) bool {
    return (strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz") && strings.ContainsAny(password, "0123456789") && strings.ContainsAny(password, "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"))
}

