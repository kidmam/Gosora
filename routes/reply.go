package routes

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"../common"
	"../common/counters"
)

// TODO: De-duplicate the upload logic
func CreateReplySubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	tid, err := strconv.Atoi(r.PostFormValue("tid"))
	if err != nil {
		return common.PreError("Failed to convert the Topic ID", w, r)
	}

	topic, err := common.Topics.Get(tid)
	if err == sql.ErrNoRows {
		return common.PreError("Couldn't find the parent topic", w, r)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	// TODO: Add hooks to make use of headerLite
	_, ferr := common.SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.CreateReply {
		return common.NoPermissions(w, r, user)
	}

	// Handle the file attachments
	// TODO: Stop duplicating this code
	if user.Perms.UploadFiles {
		files, ok := r.MultipartForm.File["upload_files"]
		if ok {
			if len(files) > 5 {
				return common.LocalError("You can't attach more than five files", w, r, user)
			}

			for _, file := range files {
				log.Print("file.Filename ", file.Filename)
				extarr := strings.Split(file.Filename, ".")
				if len(extarr) < 2 {
					return common.LocalError("Bad file", w, r, user)
				}
				ext := extarr[len(extarr)-1]

				// TODO: Can we do this without a regex?
				reg, err := regexp.Compile("[^A-Za-z0-9]+")
				if err != nil {
					return common.LocalError("Bad file extension", w, r, user)
				}
				ext = strings.ToLower(reg.ReplaceAllString(ext, ""))
				if !common.AllowedFileExts.Contains(ext) {
					return common.LocalError("You're not allowed to upload files with this extension", w, r, user)
				}

				infile, err := file.Open()
				if err != nil {
					return common.LocalError("Upload failed", w, r, user)
				}
				defer infile.Close()

				hasher := sha256.New()
				_, err = io.Copy(hasher, infile)
				if err != nil {
					return common.LocalError("Upload failed [Hashing Failed]", w, r, user)
				}
				infile.Close()

				checksum := hex.EncodeToString(hasher.Sum(nil))
				filename := checksum + "." + ext
				outfile, err := os.Create("." + "/attachs/" + filename)
				if err != nil {
					return common.LocalError("Upload failed [File Creation Failed]", w, r, user)
				}
				defer outfile.Close()

				infile, err = file.Open()
				if err != nil {
					return common.LocalError("Upload failed", w, r, user)
				}
				defer infile.Close()

				_, err = io.Copy(outfile, infile)
				if err != nil {
					return common.LocalError("Upload failed [Copy Failed]", w, r, user)
				}

				err = common.Attachments.Add(topic.ParentID, "forums", tid, "replies", user.ID, filename)
				if err != nil {
					return common.InternalError(err, w, r)
				}
			}
		}
	}

	content := common.PreparseMessage(r.PostFormValue("reply-content"))
	// TODO: Fully parse the post and put that in the parsed column
	rid, err := common.Rstore.Create(topic, content, user.LastIP, user.ID)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	reply, err := common.Rstore.Get(rid)
	if err != nil {
		return common.LocalError("Unable to load the reply", w, r, user)
	}
	if r.PostFormValue("has_poll") == "1" {
		var maxPollOptions = 10
		var pollInputItems = make(map[int]string)
		for key, values := range r.Form {
			common.DebugDetail("key: ", key)
			common.DebugDetailf("values: %+v\n", values)
			for _, value := range values {
				if strings.HasPrefix(key, "pollinputitem[") {
					halves := strings.Split(key, "[")
					if len(halves) != 2 {
						return common.LocalError("Malformed pollinputitem", w, r, user)
					}
					halves[1] = strings.TrimSuffix(halves[1], "]")

					index, err := strconv.Atoi(halves[1])
					if err != nil {
						return common.LocalError("Malformed pollinputitem", w, r, user)
					}

					// If there are duplicates, then something has gone horribly wrong, so let's ignore them, this'll likely happen during an attack
					_, exists := pollInputItems[index]
					if !exists && len(html.EscapeString(value)) != 0 {
						pollInputItems[index] = html.EscapeString(value)
						if len(pollInputItems) >= maxPollOptions {
							break
						}
					}
				}
			}
		}

		// Make sure the indices are sequential to avoid out of bounds issues
		var seqPollInputItems = make(map[int]string)
		for i := 0; i < len(pollInputItems); i++ {
			seqPollInputItems[i] = pollInputItems[i]
		}

		pollType := 0 // Basic single choice
		_, err := common.Polls.Create(reply, pollType, seqPollInputItems)
		if err != nil {
			return common.LocalError("Failed to add poll to reply", w, r, user) // TODO: Might need to be an internal error as it could leave phantom polls?
		}
	}

	err = common.Forums.UpdateLastTopic(tid, user.ID, topic.ParentID)
	if err != nil && err != sql.ErrNoRows {
		return common.InternalError(err, w, r)
	}

	common.AddActivityAndNotifyAll(user.ID, topic.CreatedBy, "reply", "topic", tid)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	http.Redirect(w, r, "/topic/"+strconv.Itoa(tid), http.StatusSeeOther)

	wcount := common.WordCount(content)
	err = user.IncreasePostStats(wcount, false)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	counters.PostCounter.Bump()
	return nil
}

// TODO: Disable stat updates in posts handled by plugin_guilds
// TODO: Update the stats after edits so that we don't under or over decrement stats during deletes
func ReplyEditSubmit(w http.ResponseWriter, r *http.Request, user common.User, srid string) common.RouteError {
	isJs := (r.PostFormValue("js") == "1")

	rid, err := strconv.Atoi(srid)
	if err != nil {
		return common.PreErrorJSQ("The provided Reply ID is not a valid number.", w, r, isJs)
	}

	reply, err := common.Rstore.Get(rid)
	if err == sql.ErrNoRows {
		return common.PreErrorJSQ("The target reply doesn't exist.", w, r, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	topic, err := reply.Topic()
	if err == sql.ErrNoRows {
		return common.PreErrorJSQ("The parent topic doesn't exist.", w, r, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	// TODO: Add hooks to make use of headerLite
	_, ferr := common.SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.EditReply {
		return common.NoPermissionsJSQ(w, r, user, isJs)
	}

	err = reply.SetPost(r.PostFormValue("edit_item"))
	if err == sql.ErrNoRows {
		return common.PreErrorJSQ("The parent topic doesn't exist.", w, r, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	if !isJs {
		http.Redirect(w, r, "/topic/"+strconv.Itoa(topic.ID)+"#reply-"+strconv.Itoa(rid), http.StatusSeeOther)
	} else {
		w.Write(successJSONBytes)
	}
	return nil
}

// TODO: Refactor this
// TODO: Disable stat updates in posts handled by plugin_guilds
func ReplyDeleteSubmit(w http.ResponseWriter, r *http.Request, user common.User, srid string) common.RouteError {
	isJs := (r.PostFormValue("isJs") == "1")

	rid, err := strconv.Atoi(srid)
	if err != nil {
		return common.PreErrorJSQ("The provided Reply ID is not a valid number.", w, r, isJs)
	}

	reply, err := common.Rstore.Get(rid)
	if err == sql.ErrNoRows {
		return common.PreErrorJSQ("The reply you tried to delete doesn't exist.", w, r, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	topic, err := common.Topics.Get(reply.ParentID)
	if err == sql.ErrNoRows {
		return common.PreErrorJSQ("The parent topic doesn't exist.", w, r, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	// TODO: Add hooks to make use of headerLite
	_, ferr := common.SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.DeleteReply {
		return common.NoPermissionsJSQ(w, r, user, isJs)
	}

	err = reply.Delete()
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	//log.Printf("Reply #%d was deleted by common.User #%d", rid, user.ID)
	if !isJs {
		http.Redirect(w, r, "/topic/"+strconv.Itoa(reply.ParentID), http.StatusSeeOther)
	} else {
		w.Write(successJSONBytes)
	}

	replyCreator, err := common.Users.Get(reply.CreatedBy)
	if err == nil {
		wcount := common.WordCount(reply.Content)
		err = replyCreator.DecreasePostStats(wcount, false)
		if err != nil {
			return common.InternalErrorJSQ(err, w, r, isJs)
		}
	} else if err != sql.ErrNoRows {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	err = common.ModLogs.Create("delete", reply.ParentID, "reply", user.LastIP, user.ID)
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}
	return nil
}

func ProfileReplyEditSubmit(w http.ResponseWriter, r *http.Request, user common.User, srid string) common.RouteError {
	isJs := (r.PostFormValue("js") == "1")

	rid, err := strconv.Atoi(srid)
	if err != nil {
		return common.LocalErrorJSQ("The provided Reply ID is not a valid number.", w, r, user, isJs)
	}

	reply, err := common.Prstore.Get(rid)
	if err == sql.ErrNoRows {
		return common.PreErrorJSQ("The target reply doesn't exist.", w, r, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	creator, err := common.Users.Get(reply.CreatedBy)
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	// ? Does the admin understand that this group perm affects this?
	if user.ID != creator.ID && !user.Perms.EditReply {
		return common.NoPermissionsJSQ(w, r, user, isJs)
	}

	err = reply.SetBody(r.PostFormValue("edit_item"))
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	if !isJs {
		http.Redirect(w, r, "/user/"+strconv.Itoa(creator.ID)+"#reply-"+strconv.Itoa(rid), http.StatusSeeOther)
	} else {
		w.Write(successJSONBytes)
	}
	return nil
}

func ProfileReplyDeleteSubmit(w http.ResponseWriter, r *http.Request, user common.User, srid string) common.RouteError {
	isJs := (r.PostFormValue("isJs") == "1")

	rid, err := strconv.Atoi(srid)
	if err != nil {
		return common.LocalErrorJSQ("The provided Reply ID is not a valid number.", w, r, user, isJs)
	}

	reply, err := common.Prstore.Get(rid)
	if err == sql.ErrNoRows {
		return common.PreErrorJSQ("The target reply doesn't exist.", w, r, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	creator, err := common.Users.Get(reply.CreatedBy)
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	if user.ID != creator.ID && !user.Perms.DeleteReply {
		return common.NoPermissionsJSQ(w, r, user, isJs)
	}

	err = reply.Delete()
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}
	//log.Printf("The profile post '%d' was deleted by common.User #%d", reply.ID, user.ID)

	if !isJs {
		//http.Redirect(w,r, "/user/" + strconv.Itoa(creator.ID), http.StatusSeeOther)
	} else {
		w.Write(successJSONBytes)
	}
	return nil
}