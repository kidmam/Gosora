package common

import (
	"database/sql"

	"github.com/Azareal/Gosora/query_gen"
)

var Prstore ProfileReplyStore

type ProfileReplyStore interface {
	Get(id int) (*ProfileReply, error)
	Create(profileID int, content string, createdBy int, ipaddress string) (id int, err error)
}

// TODO: Refactor this to stop using the global stmt store
// TODO: Add more methods to this like Create()
type SQLProfileReplyStore struct {
	get    *sql.Stmt
	create *sql.Stmt
}

func NewSQLProfileReplyStore(acc *qgen.Accumulator) (*SQLProfileReplyStore, error) {
	return &SQLProfileReplyStore{
		get:    acc.Select("users_replies").Columns("uid, content, createdBy, createdAt, lastEdit, lastEditBy, ipaddress").Where("rid = ?").Prepare(),
		create: acc.Insert("users_replies").Columns("uid, content, parsed_content, createdAt, createdBy, ipaddress").Fields("?,?,?,UTC_TIMESTAMP(),?,?").Prepare(),
	}, acc.FirstError()
}

func (store *SQLProfileReplyStore) Get(id int) (*ProfileReply, error) {
	reply := ProfileReply{ID: id}
	err := store.get.QueryRow(id).Scan(&reply.ParentID, &reply.Content, &reply.CreatedBy, &reply.CreatedAt, &reply.LastEdit, &reply.LastEditBy, &reply.IPAddress)
	return &reply, err
}

func (store *SQLProfileReplyStore) Create(profileID int, content string, createdBy int, ipaddress string) (id int, err error) {
	res, err := store.create.Exec(profileID, content, ParseMessage(content, 0, ""), createdBy, ipaddress)
	if err != nil {
		return 0, err
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Should we reload the user?
	return int(lastID), err
}
