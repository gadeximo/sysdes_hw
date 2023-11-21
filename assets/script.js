/* placeholder file for JavaScript */
const confirm_delete = (id) => {
    if(window.confirm(`Task ${id} を削除します．よろしいですか？`)) {
        location.href = `/task/delete/${id}`;
    }
}
 
const confirmWithText = (str) => {
var result = confirm(str);
  return result; // OKがクリックされた場合のみフォーム送信が続行されます
}

function goBack() {
  window.history.back();
}

document.addEventListener("DOMContentLoaded", function() {
  // ログイン状態を取得
  const isLoggedIn = checkLoginStatus();

  // ログイン状態によってボタンを表示・非表示に切り替え
  if (isLoggedIn) {
      document.getElementById("logoutButton").style.display = "block";
  } else {
      document.getElementById("loginButton").style.display = "block";
  }


  // ログイン状態を取得する関数
  function checkLoginStatus() {
      // 仮の実装。実際はCookieからuser-sessionを取得するなど。
      return document.cookie.includes("user-session");
  }

});