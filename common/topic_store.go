/*
*
*	Gosora Topic Store
*	Copyright Azareal 2017 - 2019
*
 */
package common

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"github.com/Azareal/Gosora/query_gen"
)

// TODO: Add the watchdog goroutine
// TODO: Add BulkGetMap
// TODO: Add some sort of update method
// ? - Should we add stick, lock, unstick, and unlock methods? These might be better on the Topics not the TopicStore
var Topics TopicStore
var ErrNoTitle = errors.New("This message is missing a title")
var ErrLongTitle = errors.New("The title is too long")
var ErrNoBody = errors.New("This message is missing a body")

type TopicStore interface {
	DirtyGet(id int) *Topic
	Get(id int) (*Topic, error)
	BypassGet(id int) (*Topic, error)
	BulkGetMap(ids []int) (list map[int]*Topic, err error)
	Exists(id int) bool
	Create(fid int, topicName string, content string, uid int, ipaddress string) (tid int, err error)
	AddLastTopic(item *Topic, fid int) error // unimplemented
	Reload(id int) error                     // Too much SQL logic to move into TopicCache
	// TODO: Implement these two methods
	//Replies(tid int) ([]*Reply, error)
	//RepliesRange(tid int, lower int, higher int) ([]*Reply, error)
	GlobalCount() int

	SetCache(cache TopicCache)
	GetCache() TopicCache
}

type DefaultTopicStore struct {
	cache TopicCache

	get        *sql.Stmt
	exists     *sql.Stmt
	topicCount *sql.Stmt
	create     *sql.Stmt
}

// NewDefaultTopicStore gives you a new instance of DefaultTopicStore
func NewDefaultTopicStore(cache TopicCache) (*DefaultTopicStore, error) {
	acc := qgen.NewAcc()
	if cache == nil {
		cache = NewNullTopicCache()
	}
	return &DefaultTopicStore{
		cache:      cache,
		get:        acc.Select("topics").Columns("title, content, createdBy, createdAt, lastReplyBy, lastReplyAt, lastReplyID, is_closed, sticky, parentID, ipaddress, views, postCount, likeCount, attachCount, poll, data").Where("tid = ?").Prepare(),
		exists:     acc.Select("topics").Columns("tid").Where("tid = ?").Prepare(),
		topicCount: acc.Count("topics").Prepare(),
		create:     acc.Insert("topics").Columns("parentID, title, content, parsed_content, createdAt, lastReplyAt, lastReplyBy, ipaddress, words, createdBy").Fields("?,?,?,?,UTC_TIMESTAMP(),UTC_TIMESTAMP(),?,?,?,?").Prepare(),
	}, acc.FirstError()
}

func (mts *DefaultTopicStore) DirtyGet(id int) *Topic {
	topic, err := mts.cache.Get(id)
	if err == nil {
		return topic
	}

	topic = &Topic{ID: id}
	err = mts.get.QueryRow(id).Scan(&topic.Title, &topic.Content, &topic.CreatedBy, &topic.CreatedAt, &topic.LastReplyBy, &topic.LastReplyAt, &topic.LastReplyID, &topic.IsClosed, &topic.Sticky, &topic.ParentID, &topic.IPAddress, &topic.ViewCount, &topic.PostCount, &topic.LikeCount, &topic.AttachCount, &topic.Poll, &topic.Data)
	if err == nil {
		topic.Link = BuildTopicURL(NameToSlug(topic.Title), id)
		_ = mts.cache.Add(topic)
		return topic
	}
	return BlankTopic()
}

// TODO: Log weird cache errors?
func (mts *DefaultTopicStore) Get(id int) (topic *Topic, err error) {
	topic, err = mts.cache.Get(id)
	if err == nil {
		return topic, nil
	}

	topic = &Topic{ID: id}
	err = mts.get.QueryRow(id).Scan(&topic.Title, &topic.Content, &topic.CreatedBy, &topic.CreatedAt, &topic.LastReplyBy, &topic.LastReplyAt, &topic.LastReplyID, &topic.IsClosed, &topic.Sticky, &topic.ParentID, &topic.IPAddress, &topic.ViewCount, &topic.PostCount, &topic.LikeCount, &topic.AttachCount, &topic.Poll, &topic.Data)
	if err == nil {
		topic.Link = BuildTopicURL(NameToSlug(topic.Title), id)
		_ = mts.cache.Add(topic)
	}
	return topic, err
}

// BypassGet will always bypass the cache and pull the topic directly from the database
func (mts *DefaultTopicStore) BypassGet(id int) (*Topic, error) {
	topic := &Topic{ID: id}
	err := mts.get.QueryRow(id).Scan(&topic.Title, &topic.Content, &topic.CreatedBy, &topic.CreatedAt, &topic.LastReplyBy, &topic.LastReplyAt, &topic.LastReplyID, &topic.IsClosed, &topic.Sticky, &topic.ParentID, &topic.IPAddress, &topic.ViewCount, &topic.PostCount, &topic.LikeCount, &topic.AttachCount, &topic.Poll, &topic.Data)
	topic.Link = BuildTopicURL(NameToSlug(topic.Title), id)
	return topic, err
}

// TODO: Avoid duplicating much of this logic from user_store.go
func (s *DefaultTopicStore) BulkGetMap(ids []int) (list map[int]*Topic, err error) {
	var idCount = len(ids)
	list = make(map[int]*Topic)
	if idCount == 0 {
		return list, nil
	}

	var stillHere []int
	sliceList := s.cache.BulkGet(ids)
	if len(sliceList) > 0 {
		for i, sliceItem := range sliceList {
			if sliceItem != nil {
				list[sliceItem.ID] = sliceItem
			} else {
				stillHere = append(stillHere, ids[i])
			}
		}
		ids = stillHere
	}

	// If every user is in the cache, then return immediately
	if len(ids) == 0 {
		return list, nil
	} else if len(ids) == 1 {
		topic, err := s.Get(ids[0])
		if err != nil {
			return list, err
		}
		list[topic.ID] = topic
		return list, nil
	}

	// TODO: Add a function for the qlist stuff
	var qlist string
	var idList []interface{}
	for _, id := range ids {
		idList = append(idList, strconv.Itoa(id))
		qlist += "?,"
	}
	qlist = qlist[0 : len(qlist)-1]

	rows, err := qgen.NewAcc().Select("topics").Columns("tid, title, content, createdBy, createdAt, lastReplyBy, lastReplyAt, lastReplyID, is_closed, sticky, parentID, ipaddress, views, postCount, likeCount, attachCount, poll, data").Where("tid IN(" + qlist + ")").Query(idList...)
	if err != nil {
		return list, err
	}
	defer rows.Close()

	for rows.Next() {
		topic := &Topic{}
		err := rows.Scan(&topic.ID, &topic.Title, &topic.Content, &topic.CreatedBy, &topic.CreatedAt, &topic.LastReplyBy, &topic.LastReplyAt, &topic.LastReplyID, &topic.IsClosed, &topic.Sticky, &topic.ParentID, &topic.IPAddress, &topic.ViewCount, &topic.PostCount, &topic.LikeCount, &topic.AttachCount, &topic.Poll, &topic.Data)
		if err != nil {
			return list, err
		}
		topic.Link = BuildTopicURL(NameToSlug(topic.Title), topic.ID)
		s.cache.Set(topic)
		list[topic.ID] = topic
	}
	err = rows.Err()
	if err != nil {
		return list, err
	}

	// Did we miss any topics?
	if idCount > len(list) {
		var sidList string
		for _, id := range ids {
			_, ok := list[id]
			if !ok {
				sidList += strconv.Itoa(id) + ","
			}
		}
		if sidList != "" {
			sidList = sidList[0 : len(sidList)-1]
			err = errors.New("Unable to find topics with the following IDs: " + sidList)
		}
	}

	return list, err
}

func (mts *DefaultTopicStore) Reload(id int) error {
	topic := &Topic{ID: id}
	err := mts.get.QueryRow(id).Scan(&topic.Title, &topic.Content, &topic.CreatedBy, &topic.CreatedAt, &topic.LastReplyBy, &topic.LastReplyAt, &topic.LastReplyID, &topic.IsClosed, &topic.Sticky, &topic.ParentID, &topic.IPAddress, &topic.ViewCount, &topic.PostCount, &topic.LikeCount, &topic.AttachCount, &topic.Poll, &topic.Data)
	if err == nil {
		topic.Link = BuildTopicURL(NameToSlug(topic.Title), id)
		_ = mts.cache.Set(topic)
	} else {
		_ = mts.cache.Remove(id)
	}
	TopicListThaw.Thaw()
	return err
}

func (mts *DefaultTopicStore) Exists(id int) bool {
	return mts.exists.QueryRow(id).Scan(&id) == nil
}

func (mts *DefaultTopicStore) Create(fid int, topicName string, content string, uid int, ipaddress string) (tid int, err error) {
	if topicName == "" {
		return 0, ErrNoTitle
	}
	// ? This number might be a little screwy with Unicode, but it's the only consistent thing we have, as Unicode characters can be any number of bytes in theory?
	if len(topicName) > Config.MaxTopicTitleLength {
		return 0, ErrLongTitle
	}

	parsedContent := strings.TrimSpace(ParseMessage(content, fid, "forums"))
	if parsedContent == "" {
		return 0, ErrNoBody
	}

	wcount := WordCount(content)
	// TODO: Move this statement into the topic store
	res, err := mts.create.Exec(fid, topicName, content, parsedContent, uid, ipaddress, wcount, uid)
	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(lastID), Forums.AddTopic(int(lastID), uid, fid)
}

// ? - What is this? Do we need it? Should it be in the main store interface?
func (mts *DefaultTopicStore) AddLastTopic(item *Topic, fid int) error {
	// Coming Soon...
	return nil
}

// GlobalCount returns the total number of topics on these forums
func (mts *DefaultTopicStore) GlobalCount() (tcount int) {
	err := mts.topicCount.QueryRow().Scan(&tcount)
	if err != nil {
		LogError(err)
	}
	return tcount
}

func (mts *DefaultTopicStore) SetCache(cache TopicCache) {
	mts.cache = cache
}

// TODO: We're temporarily doing this so that you can do tcache != nil in getTopicUser. Refactor it.
func (mts *DefaultTopicStore) GetCache() TopicCache {
	_, ok := mts.cache.(*NullTopicCache)
	if ok {
		return nil
	}
	return mts.cache
}
