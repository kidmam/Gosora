package main

import "log"
import "fmt"
import "strconv"
import "net/http"
import "html"
import "database/sql"
import _ "github.com/go-sql-driver/mysql"

func route_edit_topic(w http.ResponseWriter, r *http.Request) {
	user := SessionCheck(w,r)
	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form", w, r, user)
		return          
	}
	is_js := r.PostFormValue("js")
	if is_js == "" {
		is_js = "0"
	}
	
	if !user.Is_Mod && !user.Is_Admin {
		NoPermissionsJSQ(w,r,user,is_js)
		return
	}
	if user.Is_Banned {
		BannedJSQ(w,r,user,is_js)
		return
	}
	
	var tid int
	tid, err = strconv.Atoi(r.URL.Path[len("/topic/edit/submit/"):])
	if err != nil {
		LocalErrorJSQ("The provided TopicID is not a valid number.",w,r,user,is_js)
		return
	}
	
	topic_name := r.PostFormValue("topic_name")
	topic_status := r.PostFormValue("topic_status")
	
	var is_closed bool
	if topic_status == "closed" {
		is_closed = true
	} else {
		is_closed = false
	}
	
	topic_content := html.EscapeString(r.PostFormValue("topic_content"))
	_, err = edit_topic_stmt.Exec(topic_name, preparse_message(topic_content), parse_message(html.EscapeString(preparse_message(topic_content))), is_closed, tid)
	if err != nil {
		InternalErrorJSQ(err,w,r,user,is_js)
		return
	}
	
	if is_js == "0" {
		http.Redirect(w, r, "/topic/" + strconv.Itoa(tid), http.StatusSeeOther)
	} else {
		fmt.Fprintf(w,"{'success': '1'}")
	}
}

func route_delete_topic(w http.ResponseWriter, r *http.Request) {
	user := SessionCheck(w,r)
	if !user.Is_Mod && !user.Is_Admin {
		NoPermissions(w,r,user)
		return
	}
	
	tid, err := strconv.Atoi(r.URL.Path[len("/topic/delete/submit/"):])
	if err != nil {
		LocalError("The provided TopicID is not a valid number.",w,r,user)
		return
	}
	
	err = db.QueryRow("SELECT tid from topics where tid = ?", tid).Scan(&tid)
	if err == sql.ErrNoRows {
		LocalError("The topic you tried to delete doesn't exist.",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	_, err = delete_topic_stmt.Exec(tid)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	log.Print("The topic '" + strconv.Itoa(tid) + "' was deleted by User ID #" + strconv.Itoa(user.ID) + ".")
	
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func route_stick_topic(w http.ResponseWriter, r *http.Request) {
	user := SessionCheck(w,r)
	if !user.Is_Mod && !user.Is_Admin {
		NoPermissions(w,r,user)
		return
	}
	
	tid, err := strconv.Atoi(r.URL.Path[len("/topic/stick/submit/"):])
	if err != nil {
		LocalError("The provided TopicID is not a valid number.",w,r,user)
		return
	}
	
	_, err = stick_topic_stmt.Exec(tid)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	http.Redirect(w, r, "/topic/" + strconv.Itoa(tid), http.StatusSeeOther)
}

func route_unstick_topic(w http.ResponseWriter, r *http.Request) {
	user := SessionCheck(w,r)
	if !user.Is_Mod && !user.Is_Admin {
		NoPermissions(w,r,user)
		return
	}
	
	tid, err := strconv.Atoi(r.URL.Path[len("/topic/unstick/submit/"):])
	if err != nil {
		LocalError("The provided TopicID is not a valid number.",w,r,user)
		return
	}
	
	_, err = unstick_topic_stmt.Exec(tid)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	http.Redirect(w, r, "/topic/" + strconv.Itoa(tid), http.StatusSeeOther)
}

func route_reply_edit_submit(w http.ResponseWriter, r *http.Request) {
	user := SessionCheck(w,r)
	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form", w, r, user)
		return          
	}
	
	is_js := r.PostFormValue("js")
	if is_js == "" {
		is_js = "0"
	}
	
	if !user.Is_Mod && !user.Is_Admin {
		NoPermissionsJSQ(w,r,user,is_js)
		return
	}
	
	rid, err := strconv.Atoi(r.URL.Path[len("/reply/edit/submit/"):])
	if err != nil {
		LocalError("The provided Reply ID is not a valid number.",w,r,user)
		return
	}
	
	content := html.EscapeString(preparse_message(r.PostFormValue("edit_item")))
	_, err = edit_reply_stmt.Exec(content, parse_message(content), rid)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	// Get the Reply ID..
	var tid int
	err = db.QueryRow("select tid from replies where rid = ?", rid).Scan(&tid)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	if is_js == "0" {
		http.Redirect(w,r, "/topic/" + strconv.Itoa(tid) + "#reply-" + strconv.Itoa(rid), http.StatusSeeOther)
	} else {
		fmt.Fprintf(w,"{'success': '1'}")
	}
}

func route_reply_delete_submit(w http.ResponseWriter, r *http.Request) {
	user := SessionCheck(w,r)
	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form", w, r, user)
		return          
	}
	
	is_js := r.PostFormValue("is_js")
	if is_js == "" {
		is_js = "0"
	}
	
	if !user.Is_Mod && !user.Is_Admin {
		NoPermissionsJSQ(w,r,user,is_js)
		return
	}
	
	rid, err := strconv.Atoi(r.URL.Path[len("/reply/delete/submit/"):])
	if err != nil {
		LocalErrorJSQ("The provided Reply ID is not a valid number.",w,r,user,is_js)
		return
	}
	
	var tid int
	err = db.QueryRow("SELECT tid from replies where rid = ?", rid).Scan(&tid)
	if err == sql.ErrNoRows {
		LocalErrorJSQ("The reply you tried to delete doesn't exist.",w,r,user,is_js)
		return
	} else if err != nil {
		InternalErrorJSQ(err,w,r,user,is_js)
		return
	}
	
	_, err = delete_reply_stmt.Exec(rid)
	if err != nil {
		InternalErrorJSQ(err,w,r,user,is_js)
		return
	}
	log.Print("The reply '" + strconv.Itoa(rid) + "' was deleted by User ID #" + strconv.Itoa(user.ID) + ".")
	
	if is_js == "0" {
		//http.Redirect(w,r, "/topic/" + strconv.Itoa(tid), http.StatusSeeOther)
	} else {
		fmt.Fprintf(w,"{'success': '1'}")
	}
}

func route_profile_reply_edit_submit(w http.ResponseWriter, r *http.Request) {
	user := SessionCheck(w,r)
	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form", w, r, user)
		return          
	}
	
	is_js := r.PostFormValue("js")
	if is_js == "" {
		is_js = "0"
	}
	
	rid, err := strconv.Atoi(r.URL.Path[len("/profile/reply/edit/submit/"):])
	if err != nil {
		LocalError("The provided Reply ID is not a valid number.",w,r,user)
		return
	}
	
	// Get the Reply ID..
	var uid int
	err = db.QueryRow("select uid from users_replies where rid = ?", rid).Scan(&uid)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	if user.ID != uid && !user.Is_Mod && !user.Is_Admin {
		NoPermissionsJSQ(w,r,user,is_js)
		return
	}
	
	content := html.EscapeString(preparse_message(r.PostFormValue("edit_item")))
	_, err = edit_profile_reply_stmt.Exec(content, parse_message(content), rid)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	if is_js == "0" {
		http.Redirect(w,r, "/user/" + strconv.Itoa(uid) + "#reply-" + strconv.Itoa(rid), http.StatusSeeOther)
	} else {
		fmt.Fprintf(w,"{'success': '1'}")
	}
}

func route_profile_reply_delete_submit(w http.ResponseWriter, r *http.Request) {
	user := SessionCheck(w,r)
	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form", w, r, user)
		return          
	}
	
	is_js := r.PostFormValue("is_js")
	if is_js == "" {
		is_js = "0"
	}
	
	rid, err := strconv.Atoi(r.URL.Path[len("/profile/reply/delete/submit/"):])
	if err != nil {
		LocalErrorJSQ("The provided Reply ID is not a valid number.",w,r,user,is_js)
		return
	}
	
	var uid int
	err = db.QueryRow("SELECT uid from users_replies where rid = ?", rid).Scan(&uid)
	if err == sql.ErrNoRows {
		LocalErrorJSQ("The reply you tried to delete doesn't exist.",w,r,user,is_js)
		return
	} else if err != nil {
		InternalErrorJSQ(err,w,r,user,is_js)
		return
	}
	
	if user.ID != uid && !user.Is_Mod && !user.Is_Admin {
		NoPermissionsJSQ(w,r,user,is_js)
		return
	}
	
	_, err = delete_profile_reply_stmt.Exec(rid)
	if err != nil {
		InternalErrorJSQ(err,w,r,user,is_js)
		return
	}
	log.Print("The reply '" + strconv.Itoa(rid) + "' was deleted by User ID #" + strconv.Itoa(user.ID) + ".")
	
	if is_js == "0" {
		//http.Redirect(w,r, "/user/" + strconv.Itoa(uid), http.StatusSeeOther)
	} else {
		fmt.Fprintf(w,"{'success': '1'}")
	}
}

func route_ban(w http.ResponseWriter, r *http.Request) {
	user := SessionCheck(w,r)
	if !user.Is_Mod && !user.Is_Admin {
		NoPermissions(w,r,user)
		return
	}
	
	uid, err := strconv.Atoi(r.URL.Path[len("/users/ban/"):])
	if err != nil {
		LocalError("The provided User ID is not a valid number.",w,r,user)
		return
	}
	
	var uname string
	err = db.QueryRow("SELECT name from users where uid = ?", uid).Scan(&uname)
	if err == sql.ErrNoRows {
		LocalError("The user you're trying to ban no longer exists.",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	confirm_msg := "Are you sure you want to ban '" + uname + "'?"
	yousure := AreYouSure{"/users/ban/submit/" + strconv.Itoa(uid),confirm_msg}
	
	pi := Page{"Ban User","ban-user",user,tList,yousure}
	templates.ExecuteTemplate(w,"areyousure.html", pi)
}

func route_ban_submit(w http.ResponseWriter, r *http.Request) {
	user := SessionCheck(w,r)
	if !user.Is_Mod && !user.Is_Admin {
		NoPermissions(w,r,user)
		return
	}
	if r.FormValue("session") != user.Session {
		SecurityError(w,r,user)
		return
	}
	
	uid, err := strconv.Atoi(r.URL.Path[len("/users/ban/submit/"):])
	if err != nil {
		LocalError("The provided User ID is not a valid number.",w,r,user)
		return
	}
	
	var group int
	var is_super_admin bool
	err = db.QueryRow("SELECT `group`, `is_super_admin` from `users` where `uid` = ?", uid).Scan(&group, &is_super_admin)
	if err == sql.ErrNoRows {
		LocalError("The user you're trying to ban no longer exists.",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	if is_super_admin || groups[group].Is_Admin || groups[group].Is_Mod {
		LocalError("You may not ban another staff member.",w,r,user)
		return
	}
	if uid == user.ID {
		LocalError("You may not ban yourself.",w,r,user)
		return
	}
	if uid == -2 {
		LocalError("You may not ban me. Fine, I will offer up some guidance unto thee. Come to my lair, young one. /arcane-tower/",w,r,user)
		return
	}
	
	if groups[group].Is_Banned {
		LocalError("The user you're trying to unban is already banned.",w,r,user)
		return
	}
	
	_, err = change_group_stmt.Exec(4, uid)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	http.Redirect(w,r,"/users/" + strconv.Itoa(uid),http.StatusSeeOther)
}

func route_unban(w http.ResponseWriter, r *http.Request) {
	user := SessionCheck(w,r)
	if !user.Is_Mod && !user.Is_Admin {
		NoPermissions(w,r,user)
		return
	}
	
	uid, err := strconv.Atoi(r.URL.Path[len("/users/unban/"):])
	if err != nil {
		LocalError("The provided User ID is not a valid number.",w,r,user)
		return
	}
	
	var uname string
	var group int
	err = db.QueryRow("SELECT `name`, `group` from users where `uid` = ?", uid).Scan(&uname, &group)
	if err == sql.ErrNoRows {
		LocalError("The user you're trying to unban no longer exists.",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	if !groups[group].Is_Banned {
		LocalError("The user you're trying to unban isn't banned.",w,r,user)
		return
	}
	
	_, err = change_group_stmt.Exec(default_group, uid)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	http.Redirect(w,r,"/users/" + strconv.Itoa(uid),http.StatusSeeOther)
}

func route_panel_forums(w http.ResponseWriter, r *http.Request){
	user := SessionCheck(w,r)
	if !user.Is_Admin {
		NoPermissions(w,r,user)
		return
	}
	
	var forumList map[int]interface{} = make(map[int]interface{})
	currentID := 0
	
	for _, forum := range forums {
		if forum.ID > -1 {
			forumList[currentID] = forum
			currentID++
		}
	}
	
	pi := Page{"Forum Manager","panel-forums",user,forumList,0}
	templates.ExecuteTemplate(w,"panel-forums.html", pi)
}

func route_panel_forums_create_submit(w http.ResponseWriter, r *http.Request){
	user := SessionCheck(w,r)
	if !user.Is_Admin {
		NoPermissions(w,r,user)
		return
	}
	
	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form", w, r, user)
		return          
	}
	if r.FormValue("session") != user.Session {
		SecurityError(w,r,user)
		return
	}
	
	fname := r.PostFormValue("forum-name")
	res, err := create_forum_stmt.Exec(fname)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	lastId, err := res.LastInsertId()
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	forums[int(lastId)] = Forum{int(lastId),fname,true,"",0,"",0,""}
	http.Redirect(w,r,"/panel/forums/",http.StatusSeeOther)
}

func route_panel_forums_delete(w http.ResponseWriter, r *http.Request){
	user := SessionCheck(w,r)
	if !user.Is_Admin {
		NoPermissions(w,r,user)
		return
	}
	if r.FormValue("session") != user.Session {
		SecurityError(w,r,user)
		return
	}
	fid, err := strconv.Atoi(r.URL.Path[len("/panel/forums/delete/"):])
	if err != nil {
		LocalError("The provided Forum ID is not a valid number.",w,r,user)
		return
	}
	
	_, ok := forums[fid];
    if !ok {
		LocalError("The forum you're trying to delete doesn't exist.",w,r,user)
		return
	}
	
	confirm_msg := "Are you sure you want to delete the '" + forums[fid].Name + "' forum?"
	yousure := AreYouSure{"/panel/forums/delete/submit/" + strconv.Itoa(fid),confirm_msg}
	
	pi := Page{"Delete Forum","panel-forums-delete",user,tList,yousure}
	templates.ExecuteTemplate(w,"areyousure.html", pi)
}

func route_panel_forums_delete_submit(w http.ResponseWriter, r *http.Request) {
	user := SessionCheck(w,r)
	if !user.Is_Admin {
		NoPermissions(w,r,user)
		return
	}
	if r.FormValue("session") != user.Session {
		SecurityError(w,r,user)
		return
	}
	
	fid, err := strconv.Atoi(r.URL.Path[len("/panel/forums/delete/submit/"):])
	if err != nil {
		LocalError("The provided Forum ID is not a valid number.",w,r,user)
		return
	}
	
	_, ok := forums[fid];
    if !ok {
		LocalError("The forum you're trying to delete doesn't exist.",w,r,user)
		return
	}
	
	_, err = delete_forum_stmt.Exec(fid)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	// Remove this forum from the forum cache
	delete(forums,fid);
	http.Redirect(w,r,"/panel/forums/",http.StatusSeeOther)
}

func route_panel_forums_edit_submit(w http.ResponseWriter, r *http.Request) {
	user := SessionCheck(w,r)
	if !user.Is_Admin {
		NoPermissions(w,r,user)
		return
	}
	
	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form", w, r, user)
		return          
	}
	if r.FormValue("session") != user.Session {
		SecurityError(w,r,user)
		return
	}
	
	fid, err := strconv.Atoi(r.URL.Path[len("/panel/forums/edit/submit/"):])
	if err != nil {
		LocalError("The provided Forum ID is not a valid number.",w,r,user)
		return
	}
	
	forum_name := r.PostFormValue("edit_item")
	
	forum, ok := forums[fid];
    if !ok {
		LocalError("The forum you're trying to edit doesn't exist.",w,r,user)
		return
	}
	
	_, err = update_forum_stmt.Exec(forum_name, fid)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	forum.Name = forum_name
	forums[fid] = forum
	
	http.Redirect(w,r,"/panel/forums/",http.StatusSeeOther)
}
