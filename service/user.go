package service
 
import (
    "net/http"
	"crypto/sha256"
    "encoding/hex"
    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/sessions"
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

func ShowLoginPage(ctx *gin.Context){
    ctx.HTML(http.StatusOK, "login.html", gin.H{"Title": "Login"})
}

const userkey = "user"
 
func Login(ctx *gin.Context) {
    username := ctx.PostForm("username")
    password := ctx.PostForm("password")
 
    db, err := database.GetConnection()
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
 
    // ユーザの取得
    var user database.User
    err = db.Get(&user, "SELECT id, name, password FROM users WHERE name = ?", username)
    if err != nil {
        ctx.HTML(http.StatusBadRequest, "login.html", gin.H{"Title": "Login", "Username": username, "Error": "No such user"})
        return
    }
 
    // パスワードの照合
    if hex.EncodeToString(user.Password) != hex.EncodeToString(hash(password)) {
        ctx.HTML(http.StatusBadRequest, "login.html", gin.H{"Title": "Login", "Username": username, "Error": "Incorrect password"})
        return
    }
 
    // セッションの保存
    session := sessions.Default(ctx)
    session.Set(userkey, user.ID)
    session.Save()
 
    ctx.Redirect(http.StatusFound, "/list")
}

func LoginCheck(ctx *gin.Context) {
    if sessions.Default(ctx).Get(userkey) == nil {
        ctx.Redirect(http.StatusFound, "/login")
        ctx.Abort()
    } else {
        ctx.Next()
    }
}

func Logout(ctx *gin.Context) {
    session := sessions.Default(ctx)
    session.Clear()
    session.Options(sessions.Options{MaxAge: -1})
    session.Save()
    ctx.Redirect(http.StatusFound, "/")
}

func ShowAccountPage(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	var username string
	err = db.Get(&username,"SELECT name FROM users WHERE id = ?" ,userID)
	if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    ctx.HTML(http.StatusOK, "account.html", gin.H{"Title": "Account", "UserName":username })
}

func ShowRepasswordPage(ctx *gin.Context){
	userID := sessions.Default(ctx).Get("user")
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	var username string
	err = db.Get(&username,"SELECT name FROM users WHERE id = ?" ,userID)
	if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    ctx.HTML(http.StatusOK, "form_edit_password.html", gin.H{"Title": "Edit password","Username": username })
}
func EditUserPassword(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	var user database.User
    err = db.Get(&user, "SELECT id, name, password FROM users WHERE id = ?", userID)
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    username := user.Name
    // フォームデータの受け取り
    oldpassword := ctx.PostForm("oldpassword")
    newpassword := ctx.PostForm("newpassword")
	newpasswordRe:= ctx.PostForm("newpasswordRe")
    switch {
    case oldpassword == "" || newpassword == "" || newpasswordRe == "":
        ctx.HTML(http.StatusBadRequest, "form_edit_password.html", gin.H{"Title": "Edit password","Username": username, "Error": "please, fill all blank", "Oldpassword": oldpassword, "Newpassword": newpassword, "NewpasswordRe": newpasswordRe})
		return
	case newpassword != newpasswordRe:
        ctx.HTML(http.StatusBadRequest, "form_edit_password.html", gin.H{"Title": "Edit password","Username": username, "Error": "confirm password did not match ", "Oldpassword": oldpassword, "Newpassword": newpassword, "NewpasswordRe": newpasswordRe})
		return
    case !isPasswordComplex(newpassword):
        ctx.HTML(http.StatusBadRequest, "form_edit_password.html", gin.H{"Title": "Edit password","Username": username, "Error": "Password is not complex, please use at least one letter (A-Z a-z) and least one digit (0-9) and special character (!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~)", "Oldpassword": oldpassword, "Newpassword": newpassword, "NewpasswordRe": newpasswordRe})
        return
    case len(newpassword)<10:
        ctx.HTML(http.StatusBadRequest, "form_edit_password.html", gin.H{"Title": "Edit password","Username": username, "Error": "Password is short, please at least 10 characters", "Oldpassword": oldpassword, "Newpassword": newpassword, "NewpasswordRe": newpasswordRe})
        return
    }

    if hex.EncodeToString(user.Password) != hex.EncodeToString(hash(oldpassword)) {
        ctx.HTML(http.StatusBadRequest, "form_edit_password.html", gin.H{"Title": "Edit password","Username": username, "Error": "old password is not correct.", "Oldpassword": oldpassword, "Newpassword": newpassword, "NewpasswordRe": newpasswordRe})
        return
    }
	_, err = db.Exec("UPDATE users SET password = ? WHERE id = ?", hash(newpassword), userID)
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
 
    ctx.Redirect(http.StatusFound, "/user/account")
}
func ShowRenamePage(ctx *gin.Context){
    userID := sessions.Default(ctx).Get("user")
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	var username string
	err = db.Get(&username,"SELECT name FROM users WHERE id = ?" ,userID)
	if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    ctx.HTML(http.StatusOK, "form_edit_username.html", gin.H{"Title": "Edit username","Username": username })
}
func EditUsername(ctx *gin.Context) {
    userID := sessions.Default(ctx).Get("user")
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	var user database.User
    err = db.Get(&user, "SELECT id, name, password FROM users WHERE id = ?", userID)
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    newusername := ctx.PostForm("newusername")
    switch {
    case newusername == "":
        ctx.HTML(http.StatusBadRequest, "form_edit_username.html", gin.H{"Title": "Edit username","Username": user.Name, "Error": "please fill blank."})
        return
    case user.Name == newusername:
        ctx.HTML(http.StatusBadRequest, "form_edit_username.html", gin.H{"Title": "Edit username","Username": user.Name, "Newusername": newusername, "Error": "same as old name"})
        return 
    }
    var duplicate int
    err = db.Get(&duplicate, "SELECT COUNT(*) FROM users WHERE name=?", newusername)
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    if duplicate > 0 {
        ctx.HTML(http.StatusBadRequest, "form_edit_username.html", gin.H{"Title": "Edit username","Username": user.Name, "Newusername": newusername, "Error": "same as other person's name"})
        return
    }
    _, err = db.Exec("UPDATE users SET name = ? WHERE id = ?", newusername, userID)
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
 
    ctx.Redirect(http.StatusFound, "/user/account")
}
func DeleteUser(ctx *gin.Context){
    userID := sessions.Default(ctx).Get("user")
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
        ctx.Abort()
		return
	}
    tx := db.MustBegin()
    //外部キー制約、カスケード制約によって自動的にオーナーシップも削除
    _ , err = tx.Exec("DELETE FROM users WHERE id=?", userID)
    if err != nil {
        tx.Rollback()
        Error(http.StatusInternalServerError, err.Error())(ctx)
        ctx.Abort()
        return
    }
    //オーナーシップ削除後参照がなくなったタスク削除、時間が過度にかかる懸念あり
    _, err = tx.Exec("DELETE FROM tasks WHERE id NOT IN (SELECT DISTINCT task_id FROM ownership)")
    if err != nil {
        tx.Rollback()
        Error(http.StatusInternalServerError, err.Error())(ctx)
        ctx.Abort()
        return
    }
    tx.Commit()
    //セッションを消去する作業を行わないと無効なクッキーが残り続ける。
    ctx.Next()
}