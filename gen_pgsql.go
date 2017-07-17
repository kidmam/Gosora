// +build pgsql

// This file was generated by Gosora's Query Generator. Please try to avoid modifying this file, as it might change at any time.
package main

import "log"
import "database/sql"

var add_replies_to_topic_stmt *sql.Stmt
var remove_replies_from_topic_stmt *sql.Stmt
var add_topics_to_forum_stmt *sql.Stmt
var remove_topics_from_forum_stmt *sql.Stmt
var update_forum_cache_stmt *sql.Stmt
var add_likes_to_topic_stmt *sql.Stmt
var add_likes_to_reply_stmt *sql.Stmt
var edit_topic_stmt *sql.Stmt
var edit_reply_stmt *sql.Stmt
var stick_topic_stmt *sql.Stmt
var unstick_topic_stmt *sql.Stmt
var update_last_ip_stmt *sql.Stmt
var update_session_stmt *sql.Stmt
var set_password_stmt *sql.Stmt
var set_avatar_stmt *sql.Stmt
var set_username_stmt *sql.Stmt
var change_group_stmt *sql.Stmt
var activate_user_stmt *sql.Stmt
var update_user_level_stmt *sql.Stmt
var increment_user_score_stmt *sql.Stmt
var increment_user_posts_stmt *sql.Stmt
var increment_user_bigposts_stmt *sql.Stmt
var increment_user_megaposts_stmt *sql.Stmt
var increment_user_topics_stmt *sql.Stmt
var edit_profile_reply_stmt *sql.Stmt
var delete_forum_stmt *sql.Stmt
var update_forum_stmt *sql.Stmt
var update_setting_stmt *sql.Stmt
var update_plugin_stmt *sql.Stmt
var update_plugin_install_stmt *sql.Stmt
var update_theme_stmt *sql.Stmt
var update_user_stmt *sql.Stmt
var update_group_perms_stmt *sql.Stmt
var update_group_rank_stmt *sql.Stmt
var update_group_stmt *sql.Stmt
var update_email_stmt *sql.Stmt
var verify_email_stmt *sql.Stmt

func _gen_pgsql() (err error) {
	if dev.DebugMode {
		log.Print("Building the generated statements")
	}
	
	log.Print("Preparing add_replies_to_topic statement.")
	add_replies_to_topic_stmt, err = db.Prepare("UPDATE `topics` SET `postCount` = `postCount` + ?,`lastReplyAt` = UTC_TIMESTAMP() WHERE `tid` = ?")
	if err != nil {
		return err
	}
		
	log.Print("Preparing remove_replies_from_topic statement.")
	remove_replies_from_topic_stmt, err = db.Prepare("UPDATE `topics` SET `postCount` = `postCount` - ? WHERE `tid` = ?")
	if err != nil {
		return err
	}
		
	log.Print("Preparing add_topics_to_forum statement.")
	add_topics_to_forum_stmt, err = db.Prepare("UPDATE `forums` SET `topicCount` = `topicCount` + ? WHERE `fid` = ?")
	if err != nil {
		return err
	}
		
	log.Print("Preparing remove_topics_from_forum statement.")
	remove_topics_from_forum_stmt, err = db.Prepare("UPDATE `forums` SET `topicCount` = `topicCount` - ? WHERE `fid` = ?")
	if err != nil {
		return err
	}
		
	log.Print("Preparing update_forum_cache statement.")
	update_forum_cache_stmt, err = db.Prepare("UPDATE `forums` SET `lastTopic` = ?,`lastTopicID` = ?,`lastReplyer` = ?,`lastReplyerID` = ?,`lastTopicTime` = UTC_TIMESTAMP() WHERE `fid` = ?")
	if err != nil {
		return err
	}
		
	log.Print("Preparing add_likes_to_topic statement.")
	add_likes_to_topic_stmt, err = db.Prepare("UPDATE `topics` SET `likeCount` = `likeCount` + ? WHERE `tid` = ?")
	if err != nil {
		return err
	}
		
	log.Print("Preparing add_likes_to_reply statement.")
	add_likes_to_reply_stmt, err = db.Prepare("UPDATE `replies` SET `likeCount` = `likeCount` + ? WHERE `rid` = ?")
	if err != nil {
		return err
	}
		
	log.Print("Preparing edit_topic statement.")
	edit_topic_stmt, err = db.Prepare("UPDATE `topics` SET `title` = ?,`content` = ?,`parsed_content` = ?,`is_closed` = ? WHERE `tid` = ?")
	if err != nil {
		return err
	}
		
	log.Print("Preparing edit_reply statement.")
	edit_reply_stmt, err = db.Prepare("UPDATE `replies` SET `content` = ?,`parsed_content` = ? WHERE `rid` = ?")
	if err != nil {
		return err
	}
		
	log.Print("Preparing stick_topic statement.")
	stick_topic_stmt, err = db.Prepare("UPDATE `topics` SET `sticky` = 1 WHERE `tid` = ?")
	if err != nil {
		return err
	}
		
	log.Print("Preparing unstick_topic statement.")
	unstick_topic_stmt, err = db.Prepare("UPDATE `topics` SET `sticky` = 0 WHERE `tid` = ?")
	if err != nil {
		return err
	}
		
	log.Print("Preparing update_last_ip statement.")
	update_last_ip_stmt, err = db.Prepare("UPDATE `users` SET `last_ip` = ? WHERE `uid` = ?")
	if err != nil {
		return err
	}
		
	log.Print("Preparing update_session statement.")
	update_session_stmt, err = db.Prepare("UPDATE `users` SET `session` = ? WHERE `uid` = ?")
	if err != nil {
		return err
	}
		
	log.Print("Preparing set_password statement.")
	set_password_stmt, err = db.Prepare("UPDATE `users` SET `password` = ?,`salt` = ? WHERE `uid` = ?")
	if err != nil {
		return err
	}
		
	log.Print("Preparing set_avatar statement.")
	set_avatar_stmt, err = db.Prepare("UPDATE `users` SET `avatar` = ? WHERE `uid` = ?")
	if err != nil {
		return err
	}
		
	log.Print("Preparing set_username statement.")
	set_username_stmt, err = db.Prepare("UPDATE `users` SET `name` = ? WHERE `uid` = ?")
	if err != nil {
		return err
	}
		
	log.Print("Preparing change_group statement.")
	change_group_stmt, err = db.Prepare("UPDATE `users` SET `group` = ? WHERE `uid` = ?")
	if err != nil {
		return err
	}
		
	log.Print("Preparing activate_user statement.")
	activate_user_stmt, err = db.Prepare("UPDATE `users` SET `active` = 1 WHERE `uid` = ?")
	if err != nil {
		return err
	}
		
	log.Print("Preparing update_user_level statement.")
	update_user_level_stmt, err = db.Prepare("UPDATE `users` SET `level` = ? WHERE `uid` = ?")
	if err != nil {
		return err
	}
		
	log.Print("Preparing increment_user_score statement.")
	increment_user_score_stmt, err = db.Prepare("UPDATE `users` SET `score` = `score` + ? WHERE `uid` = ?")
	if err != nil {
		return err
	}
		
	log.Print("Preparing increment_user_posts statement.")
	increment_user_posts_stmt, err = db.Prepare("UPDATE `users` SET `posts` = `posts` + ? WHERE `uid` = ?")
	if err != nil {
		return err
	}
		
	log.Print("Preparing increment_user_bigposts statement.")
	increment_user_bigposts_stmt, err = db.Prepare("UPDATE `users` SET `posts` = `posts` + ?,`bigposts` = `bigposts` + ? WHERE `uid` = ?")
	if err != nil {
		return err
	}
		
	log.Print("Preparing increment_user_megaposts statement.")
	increment_user_megaposts_stmt, err = db.Prepare("UPDATE `users` SET `posts` = `posts` + ?,`bigposts` = `bigposts` + ?,`megaposts` = `megaposts` + ? WHERE `uid` = ?")
	if err != nil {
		return err
	}
		
	log.Print("Preparing increment_user_topics statement.")
	increment_user_topics_stmt, err = db.Prepare("UPDATE `users` SET `topics` = `topics` + ? WHERE `uid` = ?")
	if err != nil {
		return err
	}
		
	log.Print("Preparing edit_profile_reply statement.")
	edit_profile_reply_stmt, err = db.Prepare("UPDATE `users_replies` SET `content` = ?,`parsed_content` = ? WHERE `rid` = ?")
	if err != nil {
		return err
	}
		
	log.Print("Preparing delete_forum statement.")
	delete_forum_stmt, err = db.Prepare("UPDATE `forums` SET `name` = '',`active` = 0 WHERE `fid` = ?")
	if err != nil {
		return err
	}
		
	log.Print("Preparing update_forum statement.")
	update_forum_stmt, err = db.Prepare("UPDATE `forums` SET `name` = ?,`desc` = ?,`active` = ?,`preset` = ? WHERE `fid` = ?")
	if err != nil {
		return err
	}
		
	log.Print("Preparing update_setting statement.")
	update_setting_stmt, err = db.Prepare("UPDATE `settings` SET `content` = ? WHERE `name` = ?")
	if err != nil {
		return err
	}
		
	log.Print("Preparing update_plugin statement.")
	update_plugin_stmt, err = db.Prepare("UPDATE `plugins` SET `active` = ? WHERE `uname` = ?")
	if err != nil {
		return err
	}
		
	log.Print("Preparing update_plugin_install statement.")
	update_plugin_install_stmt, err = db.Prepare("UPDATE `plugins` SET `installed` = ? WHERE `uname` = ?")
	if err != nil {
		return err
	}
		
	log.Print("Preparing update_theme statement.")
	update_theme_stmt, err = db.Prepare("UPDATE `themes` SET `default` = ? WHERE `uname` = ?")
	if err != nil {
		return err
	}
		
	log.Print("Preparing update_user statement.")
	update_user_stmt, err = db.Prepare("UPDATE `users` SET `name` = ?,`email` = ?,`group` = ? WHERE `uid` = ?")
	if err != nil {
		return err
	}
		
	log.Print("Preparing update_group_perms statement.")
	update_group_perms_stmt, err = db.Prepare("UPDATE `users_groups` SET `permissions` = ? WHERE `gid` = ?")
	if err != nil {
		return err
	}
		
	log.Print("Preparing update_group_rank statement.")
	update_group_rank_stmt, err = db.Prepare("UPDATE `users_groups` SET `is_admin` = ?,`is_mod` = ?,`is_banned` = ? WHERE `gid` = ?")
	if err != nil {
		return err
	}
		
	log.Print("Preparing update_group statement.")
	update_group_stmt, err = db.Prepare("UPDATE `users_groups` SET `name` = ?,`tag` = ? WHERE `gid` = ?")
	if err != nil {
		return err
	}
		
	log.Print("Preparing update_email statement.")
	update_email_stmt, err = db.Prepare("UPDATE `emails` SET `email` = ?,`uid` = ?,`validated` = ?,`token` = ? WHERE `email` = ?")
	if err != nil {
		return err
	}
		
	log.Print("Preparing verify_email statement.")
	verify_email_stmt, err = db.Prepare("UPDATE `emails` SET `validated` = 1,`token` = '' WHERE `email` = ?")
	if err != nil {
		return err
	}
	
	return nil
}
