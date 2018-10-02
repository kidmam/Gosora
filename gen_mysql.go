// +build !pgsql,!mssql

/* This file was generated by Gosora's Query Generator. Please try to avoid modifying this file, as it might change at any time. */

package main

import "log"
import "database/sql"
import "./common"
//import "./query_gen/lib"

// nolint
type Stmts struct {
	forumEntryExists *sql.Stmt
	groupEntryExists *sql.Stmt
	getForumTopics *sql.Stmt
	addForumPermsToForum *sql.Stmt
	updateEmail *sql.Stmt
	setTempGroup *sql.Stmt
	bumpSync *sql.Stmt
	deleteActivityStreamMatch *sql.Stmt

	getActivityFeedByWatcher *sql.Stmt
	getActivityCountByWatcher *sql.Stmt
	todaysPostCount *sql.Stmt
	todaysTopicCount *sql.Stmt
	todaysTopicCountByForum *sql.Stmt
	todaysNewUserCount *sql.Stmt

	Mocks bool
}

// nolint
func _gen_mysql() (err error) {
	common.DebugLog("Building the generated statements")
	
	common.DebugLog("Preparing forumEntryExists statement.")
	stmts.forumEntryExists, err = db.Prepare("SELECT `fid` FROM `forums` WHERE `name` = '' ORDER BY `fid` ASC LIMIT 0,1")
	if err != nil {
		log.Print("Error in forumEntryExists statement.")
		return err
	}
		
	common.DebugLog("Preparing groupEntryExists statement.")
	stmts.groupEntryExists, err = db.Prepare("SELECT `gid` FROM `users_groups` WHERE `name` = '' ORDER BY `gid` ASC LIMIT 0,1")
	if err != nil {
		log.Print("Error in groupEntryExists statement.")
		return err
	}
		
	common.DebugLog("Preparing getForumTopics statement.")
	stmts.getForumTopics, err = db.Prepare("SELECT `topics`.`tid`, `topics`.`title`, `topics`.`content`, `topics`.`createdBy`, `topics`.`is_closed`, `topics`.`sticky`, `topics`.`createdAt`, `topics`.`lastReplyAt`, `topics`.`parentID`, `users`.`name`, `users`.`avatar` FROM `topics` LEFT JOIN `users` ON `topics`.`createdBy` = `users`.`uid`  WHERE `topics`.`parentID` = ? ORDER BY `topics`.`sticky` DESC,`topics`.`lastReplyAt` DESC,`topics`.`createdBy` DESC")
	if err != nil {
		log.Print("Error in getForumTopics statement.")
		return err
	}
		
	common.DebugLog("Preparing addForumPermsToForum statement.")
	stmts.addForumPermsToForum, err = db.Prepare("INSERT INTO `forums_permissions`(`gid`,`fid`,`preset`,`permissions`) VALUES (?,?,?,?)")
	if err != nil {
		log.Print("Error in addForumPermsToForum statement.")
		return err
	}
		
	common.DebugLog("Preparing updateEmail statement.")
	stmts.updateEmail, err = db.Prepare("UPDATE `emails` SET `email` = ?,`uid` = ?,`validated` = ?,`token` = ? WHERE `email` = ?")
	if err != nil {
		log.Print("Error in updateEmail statement.")
		return err
	}
		
	common.DebugLog("Preparing setTempGroup statement.")
	stmts.setTempGroup, err = db.Prepare("UPDATE `users` SET `temp_group` = ? WHERE `uid` = ?")
	if err != nil {
		log.Print("Error in setTempGroup statement.")
		return err
	}
		
	common.DebugLog("Preparing bumpSync statement.")
	stmts.bumpSync, err = db.Prepare("UPDATE `sync` SET `last_update` = UTC_TIMESTAMP()")
	if err != nil {
		log.Print("Error in bumpSync statement.")
		return err
	}
		
	common.DebugLog("Preparing deleteActivityStreamMatch statement.")
	stmts.deleteActivityStreamMatch, err = db.Prepare("DELETE FROM `activity_stream_matches` WHERE `watcher` = ? AND `asid` = ?")
	if err != nil {
		log.Print("Error in deleteActivityStreamMatch statement.")
		return err
	}
	
	return nil
}
