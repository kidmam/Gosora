/*
*
*	Gosora Forum Store
* 	Copyright Azareal 2017 - 2019
*
 */
package common

import (
	"database/sql"
	"errors"
	"log"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/Azareal/Gosora/query_gen"
)

var forumCreateMutex sync.Mutex
var forumPerms map[int]map[int]*ForumPerms // [gid][fid]*ForumPerms // TODO: Add an abstraction around this and make it more thread-safe
var Forums ForumStore

// ForumStore is an interface for accessing the forums and the metadata stored on them
type ForumStore interface {
	LoadForums() error
	DirtyGet(id int) *Forum
	Get(id int) (*Forum, error)
	BypassGet(id int) (*Forum, error)
	BulkGetCopy(ids []int) (forums []Forum, err error)
	Reload(id int) error // ? - Should we move this to ForumCache? It might require us to do some unnecessary casting though
	//Update(Forum) error
	Delete(id int) error
	AddTopic(tid int, uid int, fid int) error
	RemoveTopic(fid int) error
	UpdateLastTopic(tid int, uid int, fid int) error
	Exists(id int) bool
	GetAll() ([]*Forum, error)
	GetAllIDs() ([]int, error)
	GetAllVisible() ([]*Forum, error)
	GetAllVisibleIDs() ([]int, error)
	//GetChildren(parentID int, parentType string) ([]*Forum,error)
	//GetFirstChild(parentID int, parentType string) (*Forum,error)
	Create(forumName string, forumDesc string, active bool, preset string) (int, error)

	GlobalCount() int
}

type ForumCache interface {
	CacheGet(id int) (*Forum, error)
	CacheSet(forum *Forum) error
	CacheDelete(id int)
	Length() int
}

// MemoryForumStore is a struct which holds an arbitrary number of forums in memory, usually all of them, although we might introduce functionality to hold a smaller subset in memory for sites with an extremely large number of forums
type MemoryForumStore struct {
	forums    sync.Map     // map[int]*Forum
	forumView atomic.Value // []*Forum

	get          *sql.Stmt
	getAll       *sql.Stmt
	delete       *sql.Stmt
	create       *sql.Stmt
	count        *sql.Stmt
	updateCache  *sql.Stmt
	addTopics    *sql.Stmt
	removeTopics *sql.Stmt
}

// NewMemoryForumStore gives you a new instance of MemoryForumStore
func NewMemoryForumStore() (*MemoryForumStore, error) {
	acc := qgen.NewAcc()
	// TODO: Do a proper delete
	return &MemoryForumStore{
		get:          acc.Select("forums").Columns("name, desc, active, preset, parentID, parentType, topicCount, lastTopicID, lastReplyerID").Where("fid = ?").Prepare(),
		getAll:       acc.Select("forums").Columns("fid, name, desc, active, preset, parentID, parentType, topicCount, lastTopicID, lastReplyerID").Orderby("fid ASC").Prepare(),
		delete:       acc.Update("forums").Set("name= '', active = 0").Where("fid = ?").Prepare(),
		create:       acc.Insert("forums").Columns("name, desc, active, preset").Fields("?,?,?,?").Prepare(),
		count:        acc.Count("forums").Where("name != ''").Prepare(),
		updateCache:  acc.Update("forums").Set("lastTopicID = ?, lastReplyerID = ?").Where("fid = ?").Prepare(),
		addTopics:    acc.Update("forums").Set("topicCount = topicCount + ?").Where("fid = ?").Prepare(),
		removeTopics: acc.Update("forums").Set("topicCount = topicCount - ?").Where("fid = ?").Prepare(),
	}, acc.FirstError()
}

// TODO: Add support for subforums
func (mfs *MemoryForumStore) LoadForums() error {
	var forumView []*Forum
	addForum := func(forum *Forum) {
		mfs.forums.Store(forum.ID, forum)
		if forum.Active && forum.Name != "" && forum.ParentType == "" {
			forumView = append(forumView, forum)
		}
	}

	rows, err := mfs.getAll.Query()
	if err != nil {
		return err
	}
	defer rows.Close()

	var i = 0
	for ; rows.Next(); i++ {
		forum := &Forum{ID: 0, Active: true, Preset: "all"}
		err = rows.Scan(&forum.ID, &forum.Name, &forum.Desc, &forum.Active, &forum.Preset, &forum.ParentID, &forum.ParentType, &forum.TopicCount, &forum.LastTopicID, &forum.LastReplyerID)
		if err != nil {
			return err
		}

		if forum.Name == "" {
			DebugLog("Adding a placeholder forum")
		} else {
			log.Printf("Adding the '%s' forum", forum.Name)
		}

		forum.Link = BuildForumURL(NameToSlug(forum.Name), forum.ID)
		forum.LastTopic = Topics.DirtyGet(forum.LastTopicID)
		forum.LastReplyer = Users.DirtyGet(forum.LastReplyerID)

		addForum(forum)
	}
	mfs.forumView.Store(forumView)
	TopicListThaw.Thaw()
	return rows.Err()
}

// TODO: Hide social groups too
// ? - Will this be hit a lot by plugin_guilds?
func (mfs *MemoryForumStore) rebuildView() {
	var forumView []*Forum
	mfs.forums.Range(func(_ interface{}, value interface{}) bool {
		forum := value.(*Forum)
		// ? - ParentType blank means that it doesn't have a parent
		if forum.Active && forum.Name != "" && forum.ParentType == "" {
			forumView = append(forumView, forum)
		}
		return true
	})
	sort.Sort(SortForum(forumView))
	mfs.forumView.Store(forumView)
}

func (mfs *MemoryForumStore) DirtyGet(id int) *Forum {
	fint, ok := mfs.forums.Load(id)
	if !ok || fint.(*Forum).Name == "" {
		return &Forum{ID: -1, Name: ""}
	}
	return fint.(*Forum)
}

func (mfs *MemoryForumStore) CacheGet(id int) (*Forum, error) {
	fint, ok := mfs.forums.Load(id)
	if !ok || fint.(*Forum).Name == "" {
		return nil, ErrNoRows
	}
	return fint.(*Forum), nil
}

func (mfs *MemoryForumStore) Get(id int) (*Forum, error) {
	fint, ok := mfs.forums.Load(id)
	if !ok || fint.(*Forum).Name == "" {
		var forum = &Forum{ID: id}
		err := mfs.get.QueryRow(id).Scan(&forum.Name, &forum.Desc, &forum.Active, &forum.Preset, &forum.ParentID, &forum.ParentType, &forum.TopicCount, &forum.LastTopicID, &forum.LastReplyerID)
		if err != nil {
			return forum, err
		}

		forum.Link = BuildForumURL(NameToSlug(forum.Name), forum.ID)
		forum.LastTopic = Topics.DirtyGet(forum.LastTopicID)
		forum.LastReplyer = Users.DirtyGet(forum.LastReplyerID)

		mfs.CacheSet(forum)
		return forum, err
	}
	return fint.(*Forum), nil
}

func (mfs *MemoryForumStore) BypassGet(id int) (*Forum, error) {
	var forum = &Forum{ID: id}
	err := mfs.get.QueryRow(id).Scan(&forum.Name, &forum.Desc, &forum.Active, &forum.Preset, &forum.ParentID, &forum.ParentType, &forum.TopicCount, &forum.LastTopicID, &forum.LastReplyerID)
	if err != nil {
		return nil, err
	}

	forum.Link = BuildForumURL(NameToSlug(forum.Name), forum.ID)
	forum.LastTopic = Topics.DirtyGet(forum.LastTopicID)
	forum.LastReplyer = Users.DirtyGet(forum.LastReplyerID)
	TopicListThaw.Thaw()

	return forum, err
}

// TODO: Optimise this
func (mfs *MemoryForumStore) BulkGetCopy(ids []int) (forums []Forum, err error) {
	forums = make([]Forum, len(ids))
	for i, id := range ids {
		forum, err := mfs.Get(id)
		if err != nil {
			return nil, err
		}
		forums[i] = forum.Copy()
	}
	return forums, nil
}

func (mfs *MemoryForumStore) Reload(id int) error {
	var forum = &Forum{ID: id}
	err := mfs.get.QueryRow(id).Scan(&forum.Name, &forum.Desc, &forum.Active, &forum.Preset, &forum.ParentID, &forum.ParentType, &forum.TopicCount, &forum.LastTopicID, &forum.LastReplyerID)
	if err != nil {
		return err
	}
	forum.Link = BuildForumURL(NameToSlug(forum.Name), forum.ID)
	forum.LastTopic = Topics.DirtyGet(forum.LastTopicID)
	forum.LastReplyer = Users.DirtyGet(forum.LastReplyerID)

	mfs.CacheSet(forum)
	TopicListThaw.Thaw()
	return nil
}

func (mfs *MemoryForumStore) CacheSet(forum *Forum) error {
	mfs.forums.Store(forum.ID, forum)
	mfs.rebuildView()
	return nil
}

// ! Has a randomised order
func (mfs *MemoryForumStore) GetAll() (forumView []*Forum, err error) {
	mfs.forums.Range(func(_ interface{}, value interface{}) bool {
		forumView = append(forumView, value.(*Forum))
		return true
	})
	sort.Sort(SortForum(forumView))
	return forumView, nil
}

// ? - Can we optimise the sorting?
func (mfs *MemoryForumStore) GetAllIDs() (ids []int, err error) {
	mfs.forums.Range(func(_ interface{}, value interface{}) bool {
		ids = append(ids, value.(*Forum).ID)
		return true
	})
	sort.Ints(ids)
	return ids, nil
}

func (mfs *MemoryForumStore) GetAllVisible() (forumView []*Forum, err error) {
	forumView = mfs.forumView.Load().([]*Forum)
	return forumView, nil
}

func (mfs *MemoryForumStore) GetAllVisibleIDs() ([]int, error) {
	forumView := mfs.forumView.Load().([]*Forum)
	var ids = make([]int, len(forumView))
	for i := 0; i < len(forumView); i++ {
		ids[i] = forumView[i].ID
	}
	return ids, nil
}

// TODO: Implement sub-forums.
/*func (mfs *MemoryForumStore) GetChildren(parentID int, parentType string) ([]*Forum,error) {
	return nil, nil
}
func (mfs *MemoryForumStore) GetFirstChild(parentID int, parentType string) (*Forum,error) {
	return nil, nil
}*/

// TODO: Add a query for this rather than hitting cache
func (mfs *MemoryForumStore) Exists(id int) bool {
	forum, ok := mfs.forums.Load(id)
	return ok && forum.(*Forum).Name != ""
}

// TODO: Batch deletions with name blanking? Is this necessary?
func (mfs *MemoryForumStore) CacheDelete(id int) {
	mfs.forums.Delete(id)
	mfs.rebuildView()
}

// TODO: Add a hook to allow plugin_guilds to detect when one of it's forums has just been deleted?
func (mfs *MemoryForumStore) Delete(id int) error {
	if id == ReportForumID {
		return errors.New("You cannot delete the Reports forum")
	}
	_, err := mfs.delete.Exec(id)
	mfs.CacheDelete(id)
	TopicListThaw.Thaw()
	return err
}

func (mfs *MemoryForumStore) AddTopic(tid int, uid int, fid int) error {
	_, err := mfs.updateCache.Exec(tid, uid, fid)
	if err != nil {
		return err
	}
	_, err = mfs.addTopics.Exec(1, fid)
	if err != nil {
		return err
	}
	// TODO: Bypass the database and update this with a lock or an unsafe atomic swap
	return mfs.Reload(fid)
}

// TODO: Update the forum cache with the latest topic
func (mfs *MemoryForumStore) RemoveTopic(fid int) error {
	_, err := mfs.removeTopics.Exec(1, fid)
	if err != nil {
		return err
	}
	// TODO: Bypass the database and update this with a lock or an unsafe atomic swap
	mfs.Reload(fid)
	return nil
}

// DEPRECATED. forum.Update() will be the way to do this in the future, once it's completed
// TODO: Have a pointer to the last topic rather than storing it on the forum itself
func (mfs *MemoryForumStore) UpdateLastTopic(tid int, uid int, fid int) error {
	_, err := mfs.updateCache.Exec(tid, uid, fid)
	if err != nil {
		return err
	}
	// TODO: Bypass the database and update this with a lock or an unsafe atomic swap
	return mfs.Reload(fid)
}

func (mfs *MemoryForumStore) Create(forumName string, forumDesc string, active bool, preset string) (int, error) {
	forumCreateMutex.Lock()
	res, err := mfs.create.Exec(forumName, forumDesc, active, preset)
	if err != nil {
		return 0, err
	}

	fid64, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	fid := int(fid64)

	err = mfs.Reload(fid)
	if err != nil {
		return 0, err
	}

	PermmapToQuery(PresetToPermmap(preset), fid)
	forumCreateMutex.Unlock()
	return fid, nil
}

// ! Might be slightly inaccurate, if the sync.Map is constantly shifting and churning, but it'll stabilise eventually. Also, slow. Don't use this on every request x.x
// Length returns the number of forums in the memory cache
func (mfs *MemoryForumStore) Length() (length int) {
	mfs.forums.Range(func(_ interface{}, value interface{}) bool {
		length++
		return true
	})
	return length
}

// TODO: Get the total count of forums in the forum store rather than doing a heavy query for this?
// GlobalCount returns the total number of forums
func (mfs *MemoryForumStore) GlobalCount() (fcount int) {
	err := mfs.count.QueryRow().Scan(&fcount)
	if err != nil {
		LogError(err)
	}
	return fcount
}

// TODO: Work on SqlForumStore

// TODO: Work on the NullForumStore
