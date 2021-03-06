// +build !pgsql,!mssql

/*
*
*	Gosora MySQL Interface
*	Copyright Azareal 2016 - 2019
*
 */
package main

import (
	"log"

	"github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/query_gen"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
)

var dbCollation = "utf8mb4_general_ci"

func init() {
	dbAdapter = "mysql"
	_initDatabase = initMySQL
}

func initMySQL() (err error) {
	err = qgen.Builder.Init("mysql", map[string]string{
		"host":      common.DbConfig.Host,
		"port":      common.DbConfig.Port,
		"name":      common.DbConfig.Dbname,
		"username":  common.DbConfig.Username,
		"password":  common.DbConfig.Password,
		"collation": dbCollation,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	// Set the number of max open connections
	db = qgen.Builder.GetConn()
	db.SetMaxOpenConns(64)
	db.SetMaxIdleConns(32)

	// Only hold connections open for five seconds to avoid accumulating a large number of stale connections
	//db.SetConnMaxLifetime(5 * time.Second)

	// Build the generated prepared statements, we are going to slowly move the queries over to the query generator rather than writing them all by hand, this'll make it easier for us to implement database adapters for other databases like PostgreSQL, MSSQL, SQlite, etc.
	err = _gen_mysql()
	if err != nil {
		return errors.WithStack(err)
	}

	// TODO: Is there a less noisy way of doing this for tests?
	log.Print("Preparing getActivityFeedByWatcher statement.")
	stmts.getActivityFeedByWatcher, err = db.Prepare("SELECT activity_stream_matches.asid, activity_stream.actor, activity_stream.targetUser, activity_stream.event, activity_stream.elementType, activity_stream.elementID FROM `activity_stream_matches` INNER JOIN `activity_stream` ON activity_stream_matches.asid = activity_stream.asid AND activity_stream_matches.watcher != activity_stream.actor WHERE `watcher` = ? ORDER BY activity_stream.asid DESC LIMIT 16")
	if err != nil {
		return errors.WithStack(err)
	}

	log.Print("Preparing getActivityCountByWatcher statement.")
	stmts.getActivityCountByWatcher, err = db.Prepare("SELECT count(*) FROM `activity_stream_matches` INNER JOIN `activity_stream` ON activity_stream_matches.asid = activity_stream.asid AND activity_stream_matches.watcher != activity_stream.actor WHERE `watcher` = ?")
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
