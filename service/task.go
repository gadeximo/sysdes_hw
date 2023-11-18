package service

import (
    //"log"
    "database/sql"
    "time"
	"net/http"
	"strconv"
	"fmt"
	"github.com/gin-gonic/gin"
    "github.com/gin-contrib/sessions"
	database "todolist.go/db"
)

// TaskList renders list of tasks in DB
func TaskList(ctx *gin.Context) {
    userID := sessions.Default(ctx).Get("user")
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
    // Get query parameter
    kw := ctx.Query("kw")
    isDoneQueryStr := ctx.Query("is_done") //"t" or "f" or ""
    sortCriterion := ctx.Query("sortCriterion")
    if sortCriterion == "" {
        sortCriterion = "createdNew"
    }

    page, err := strconv.Atoi(ctx.Query("page"))
	if err != nil {
		page = 0
	} 
    orderedby := "tasks.created_at DESC"
    switch sortCriterion {
    case "deadlineNear":
        orderedby = "tasks.deadline ASC"
    case "deadlineFar":
        orderedby = "tasks.deadline DESC"
    case "createdNew":
        orderedby = "tasks.created_at DESC"
    case "createdOld":
        orderedby = "tasks.created_at ASC"
    }

	// Get tasks in DB
	var tasks []database.Task
    var maxpage int
    var err1 error
    var err2 error
    parPage := 10
	switch {
    case isDoneQueryStr == "":
        err1 = db.Select(&tasks, "SELECT id, title, created_at, deadline ,is_done ,comment FROM tasks INNER JOIN ownership ON task_id = id WHERE user_id = ? AND title LIKE ? ORDER BY "+ orderedby+" LIMIT 10 OFFSET ?" ,userID, "%" + kw + "%", page*10)
        err2 = db.Get(&maxpage, "SELECT COUNT(*) FROM tasks INNER JOIN ownership ON task_id = id WHERE user_id = ? AND title LIKE ?", userID, "%" + kw + "%")
    default:
        isDoneQuery := (isDoneQueryStr=="t") 
        err1 = db.Select(&tasks, "SELECT id, title, created_at, deadline,is_done , comment FROM tasks INNER JOIN ownership ON task_id = id WHERE user_id = ? AND title LIKE ? AND is_done = ? ORDER BY "+ orderedby+" LIMIT 10 OFFSET ?",userID, "%" + kw + "%" , isDoneQuery, page*10)
        err2 = db.Get(&maxpage, "SELECT COUNT(*) FROM tasks INNER JOIN ownership ON task_id = id WHERE user_id = ? AND title LIKE ? AND is_done = ?", userID, "%" + kw + "%", isDoneQuery)
    }
    maxpage = (maxpage / parPage)
    pages := []int{}
	for i := 0; i <= maxpage; i++ {
		pages = append(pages, i)
	}
    if err1 != nil  {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    if err2 != nil  {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    currentTime := time.Now()

	// タスクの締め切りまでの残り日数を計算して代入
	for i := range tasks {
		daysLeft := int(tasks[i].Deadline.Sub(currentTime).Hours() / 24)
		tasks[i].DaysLeft = daysLeft
	}

	// Render tasks
	ctx.HTML(http.StatusOK, "task_list.html", gin.H{"Title": "Task list", "Tasks": tasks ,"Kw": kw, "IsDoneQuery": isDoneQueryStr, "SortCriteroin": sortCriterion, "Page": page ,"Pages": pages})
}

// ShowTask renders a task with given ID
func ShowTask(ctx *gin.Context) {
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// parse ID given as a parameter
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	// Get a task with given ID
	var task database.Task
	err = db.Get(&task, "SELECT * FROM tasks WHERE id=?", id) // Use DB#Get for one entry
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	// Render task
	//ctx.String(http.StatusOK, task.Title)  // Modify it!!
	ctx.HTML(http.StatusOK, "task.html", task)
}
//return newtask form page
func NewTaskForm(ctx *gin.Context) {
    ctx.HTML(http.StatusOK, "form_new_task.html", gin.H{"Title": "Task registration"})
}

//process for newtask form 
func RegisterTask(ctx *gin.Context) {
    userID := sessions.Default(ctx).Get("user")
    // Get task title
    title, exist := ctx.GetPostForm("title")
    if !exist {
        Error(http.StatusBadRequest, "No title is given")(ctx)
        return
    }
    deadlineStr, _ := ctx.GetPostForm("deadline")
    layout := "2006-01-02T15:04" // 入力文字列のフォーマット
	deadline, deadlineErr := time.Parse(layout, deadlineStr)
	comment, commentExist := ctx.GetPostForm("comment")
	if !commentExist || comment=="" {
		comment = "未記入"
	}
    
    // Get DB connection
    db, err := database.GetConnection()
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    // Create new data with given title on DB
    tx := db.MustBegin()
    var result sql.Result
    var Execerr error
    if deadlineErr != nil{
        result, Execerr = tx.Exec("INSERT INTO tasks (title, comment) VALUES (? , ?)", title , comment)
    } else {
        result, Execerr = tx.Exec("INSERT INTO tasks (title, comment, deadline) VALUES (? , ? , ?)", title , comment, deadline)
    }
    
    if Execerr != nil {
        tx.Rollback()
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    // Render status
    taskID, err := result.LastInsertId()
    if err != nil {
        tx.Rollback()
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    _, err = tx.Exec("INSERT INTO ownership (user_id, task_id) VALUES (?, ?)", userID, taskID)
    if err != nil {
        tx.Rollback()
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    tx.Commit()
    ctx.Redirect(http.StatusFound, fmt.Sprintf("/task/%d", taskID))
}

//return editTask page
func EditTaskForm(ctx *gin.Context) {
    // ID の取得
    id, err := strconv.Atoi(ctx.Param("id"))
    if err != nil {
        Error(http.StatusBadRequest, err.Error())(ctx)
        return
    }
    // Get DB connection
    db, err := database.GetConnection()
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    // Get target task
    var task database.Task
    err = db.Get(&task, "SELECT * FROM tasks WHERE id=?", id)
    if err != nil {
        Error(http.StatusBadRequest, err.Error())(ctx)
        return
    }
    deadlineStr := task.Deadline.Format("2006-01-02T15:04")
    // Render edit form
    ctx.HTML(http.StatusOK, "form_edit_task.html",
        gin.H{"Title": fmt.Sprintf("Edit task %d", task.ID), "Task": task, "Deadline": deadlineStr})
}

//
func UpdateTask(ctx *gin.Context){
	title, exist := ctx.GetPostForm("title")
    if !exist {
        Error(http.StatusBadRequest, "No title is given")(ctx)
        return
    }
	strIsDone, existIsDone := ctx.GetPostForm("is_done")
	if !existIsDone {
        Error(http.StatusBadRequest, "No is_done is given")(ctx)
        return
    }
	isDone, err :=strconv.ParseBool(strIsDone)
	comment, commentExist := ctx.GetPostForm("comment")
	if !commentExist || comment=="" {
		comment = "未記入"
	}
    deadlineStr, existDeadlineStr := ctx.GetPostForm("deadline")
    if !existDeadlineStr{
        Error(http.StatusBadRequest, "No deadline is given")(ctx)
        return
    }
    deadline, deadlineErr := time.Parse("2006-01-02T15:04", deadlineStr)
    if deadlineErr !=nil{
        deadline, _ = time.Parse("2006-01-02T15:04", "2000-01-01T00:00")
    }
	db, err := database.GetConnection()
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
	id, err := strconv.Atoi(ctx.Param("id"))
    if err != nil {
        Error(http.StatusBadRequest, err.Error())(ctx)
        return
    }
	_, err = db.Exec("UPDATE tasks SET title = ?, deadline = ?,is_done = ?, comment = ? WHERE id = ?", title ,deadline ,isDone, comment, id)
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
	path := "/task/"+ ctx.Param("id") // デフォルトではタスク一覧ページへ戻る
    ctx.Redirect(http.StatusFound, path)

}
//削除ボタン押された時飛ぶgetのルーティング
func DeleteTask(ctx *gin.Context) {
    userID := sessions.Default(ctx).Get("user")
    // ID の取得
    id, err := strconv.Atoi(ctx.Param("id"))
    if err != nil {
        Error(http.StatusBadRequest, err.Error())(ctx)
        return
    }
    // Get DB connection
    db, err := database.GetConnection()
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    //ownershipを削除して参照がなくなったタスクを削除する。
    tx := db.MustBegin()
    _, err = tx.Exec("DELETE FROM ownership WHERE user_id = ? AND task_id = ?", userID, id)
    if err != nil {
        tx.Rollback()
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    _, err = tx.Exec("DELETE FROM tasks WHERE id = ? AND NOT EXISTS (SELECT 1 FROM ownership WHERE task_id = ?)" , id, id)
    if err != nil {
        tx.Rollback()
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    tx.Commit()
    // Redirect to /list
    ctx.Redirect(http.StatusFound, "/list")
}

func TaskIDandUserCheck(ctx *gin.Context){
    userID := sessions.Default(ctx).Get("user")
    db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// parse ID given as a parameter
	taskid, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
    var idAndUserCount int
    err = db.Get(&idAndUserCount, "SELECT COUNT(*) FROM ownership WHERE user_id = ? AND task_id = ?;", userID, taskid)
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    if idAndUserCount == 0 {
        Error(http.StatusForbidden, "不正アクセス")(ctx)
        ctx.Abort()
    } else {
        ctx.Next()
    }
}

func ShowSharePage(ctx *gin.Context){
    taskID, err := strconv.Atoi(ctx.Param("id"))
    if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
    ctx.HTML(http.StatusOK, "form_share_task.html", gin.H{"Title": "share task","ID": taskID })
}

func Sharetask(ctx *gin.Context) {
    userID := sessions.Default(ctx).Get("user")
    taskID, err := strconv.Atoi(ctx.Param("id"))
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
    shareusername := ctx.PostForm("shareusername")
    
    var shareuserID []uint64
    err = db.Select(&shareuserID, "SELECT id FROM users WHERE name=?", shareusername)
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    if len(shareuserID) < 1 {
        ctx.HTML(http.StatusBadRequest, "form_share_task.html", gin.H{"ID": taskID ,"Title": "share task","Shareusername": shareusername, "Error": "そのユーザーは存在しません"})
        return
    } else if shareuserID[0] == userID {
        ctx.HTML(http.StatusBadRequest, "form_share_task.html", gin.H{"ID": taskID ,"Title": "share task","Shareusername": shareusername, "Error": "自身のユーザーネームは許可されません"})
        return
    }
    var duplicate int
    err = db.Get(&duplicate, "SELECT COUNT(*) FROM ownership WHERE task_id = ? AND user_id = ?", taskID, shareuserID[0])
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }

    if duplicate > 0 {
        ctx.HTML(http.StatusBadRequest, "form_share_task.html", gin.H{"ID": taskID ,"Title": "share task","Shareusername": shareusername, "Error": "共有済みです"})
        return
    }

    _, err = db.Exec("INSERT INTO ownership (task_id, user_id) VALUES (? , ?)", taskID, shareuserID[0])
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    path := "/task/"+ ctx.Param("id") // デフォルトではタスク一覧ページへ戻る
    ctx.Redirect(http.StatusFound, path)
    
}