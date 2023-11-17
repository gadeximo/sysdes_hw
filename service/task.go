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
	switch {
    case isDoneQueryStr == "":
        err = db.Select(&tasks, "SELECT id, title, created_at, deadline ,is_done ,comment FROM tasks INNER JOIN ownership ON task_id = id WHERE user_id = ? AND title LIKE ? ORDER BY "+ orderedby ,userID, "%" + kw + "%")
    default:
        isDoneQuery := (isDoneQueryStr=="t") 
        err = db.Select(&tasks, "SELECT id, title, created_at, deadline,is_done , comment FROM tasks INNER JOIN ownership ON task_id = id WHERE user_id = ? AND title LIKE ? AND is_done = ? ORDER BY "+ orderedby,userID, "%" + kw + "%" , isDoneQuery )
    }
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }

	// Render tasks
	ctx.HTML(http.StatusOK, "task_list.html", gin.H{"Title": "Task list", "Tasks": tasks ,"Kw": kw, "IsDoneQuery": isDoneQueryStr, "SortCriteroin": sortCriterion})
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
        result, Execerr = db.Exec("INSERT INTO tasks (title, comment) VALUES (? , ?)", title , comment)
    } else {
        result, Execerr = db.Exec("INSERT INTO tasks (title, comment, deadline) VALUES (? , ? , ?)", title , comment, deadline)
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
    // Delete the task from DB オーナーシップもカスケード制約により削除される。こちらをタスク削除ではなくオーナシップのみの削除にすればタスク共有の時皆から一気にタスクが消えてしまうことがなくなる。
    _, err = db.Exec("DELETE FROM tasks WHERE id=?", id)
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
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