/* placeholder file for JavaScript */
const confirm_delete = (id) => {
    if(window.confirm(`Task ${id} を削除します．よろしいですか？`)) {
        location.href = `/task/delete/${id}`;
    }
}
 
const confirm_update = (id) => {
var result = confirm("送信してもよろしいですか？");
  return result; // OKがクリックされた場合のみフォーム送信が続行されます
}

function goBack() {
  window.history.back();
}