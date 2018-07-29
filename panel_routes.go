package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"./common"
	"github.com/Azareal/gopsutil/mem"
)

// We're trying to reduce the amount of boilerplate in here, so I added these two functions, they might wind up circulating outside this file in the future
func panelSuccessRedirect(dest string, w http.ResponseWriter, r *http.Request, isJs bool) common.RouteError {
	if !isJs {
		http.Redirect(w, r, dest, http.StatusSeeOther)
	} else {
		w.Write(successJSONBytes)
	}
	return nil
}
func panelRenderTemplate(tmplName string, w http.ResponseWriter, r *http.Request, user common.User, pi interface{}) common.RouteError {
	if common.RunPreRenderHook("pre_render_"+tmplName, w, r, &user, pi) {
		return nil
	}
	err := common.Templates.ExecuteTemplate(w, tmplName+".html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

func routePanelDashboard(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	header, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	header.Title = common.GetTitlePhrase("panel_dashboard")

	// We won't calculate this on the spot anymore, as the system doesn't seem to like it if we do multiple fetches simultaneously. Should we constantly calculate this on a background thread? Perhaps, the watchdog to scale back heavy features under load? One plus side is that we'd get immediate CPU percentages here instead of waiting it to kick in with WebSockets
	var cpustr = "Unknown"
	var cpuColour string

	lessThanSwitch := func(number int, lowerBound int, midBound int) string {
		switch {
		case number < lowerBound:
			return "stat_green"
		case number < midBound:
			return "stat_orange"
		}
		return "stat_red"
	}

	var ramstr, ramColour string
	memres, err := mem.VirtualMemory()
	if err != nil {
		ramstr = "Unknown"
	} else {
		totalCount, totalUnit := common.ConvertByteUnit(float64(memres.Total))
		usedCount := common.ConvertByteInUnit(float64(memres.Total-memres.Available), totalUnit)

		// Round totals with .9s up, it's how most people see it anyway. Floats are notoriously imprecise, so do it off 0.85
		var totstr string
		if (totalCount - float64(int(totalCount))) > 0.85 {
			usedCount += 1.0 - (totalCount - float64(int(totalCount)))
			totstr = strconv.Itoa(int(totalCount) + 1)
		} else {
			totstr = fmt.Sprintf("%.1f", totalCount)
		}

		if usedCount > totalCount {
			usedCount = totalCount
		}
		ramstr = fmt.Sprintf("%.1f", usedCount) + " / " + totstr + totalUnit

		ramperc := ((memres.Total - memres.Available) * 100) / memres.Total
		ramColour = lessThanSwitch(int(ramperc), 50, 75)
	}

	greaterThanSwitch := func(number int, lowerBound int, midBound int) string {
		switch {
		case number > midBound:
			return "stat_green"
		case number > lowerBound:
			return "stat_orange"
		}
		return "stat_red"
	}

	// TODO: Add a stat store for this?
	var intErr error
	var extractStat = func(stmt *sql.Stmt, args ...interface{}) (stat int) {
		err := stmt.QueryRow(args...).Scan(&stat)
		if err != nil && err != ErrNoRows {
			intErr = err
		}
		return stat
	}

	var postCount = extractStat(stmts.todaysPostCount)
	var postInterval = "day"
	var postColour = greaterThanSwitch(postCount, 5, 25)

	var topicCount = extractStat(stmts.todaysTopicCount)
	var topicInterval = "day"
	var topicColour = greaterThanSwitch(topicCount, 0, 8)

	var reportCount = extractStat(stmts.todaysTopicCountByForum, common.ReportForumID)
	var reportInterval = "week"

	var newUserCount = extractStat(stmts.todaysNewUserCount)
	var newUserInterval = "week"

	// Did any of the extractStats fail?
	if intErr != nil {
		return common.InternalError(intErr, w, r)
	}

	// TODO: Localise these
	var gridElements = []common.GridElement{
		// TODO: Implement a check for new versions of Gosora
		//common.GridElement{"dash-version", "v" + version.String(), 0, "grid_istat stat_green", "", "", "Gosora is up-to-date :)"},
		common.GridElement{"dash-version", "v" + version.String(), 0, "grid_istat", "", "", ""},

		common.GridElement{"dash-cpu", "CPU: " + cpustr, 1, "grid_istat " + cpuColour, "", "", "The global CPU usage of this server"},
		common.GridElement{"dash-ram", "RAM: " + ramstr, 2, "grid_istat " + ramColour, "", "", "The global RAM usage of this server"},
	}
	var addElement = func(element common.GridElement) {
		gridElements = append(gridElements, element)
	}

	if common.EnableWebsockets {
		uonline := common.WsHub.UserCount()
		gonline := common.WsHub.GuestCount()
		totonline := uonline + gonline
		reqCount := 0

		var onlineColour = greaterThanSwitch(totonline, 3, 10)
		var onlineGuestsColour = greaterThanSwitch(gonline, 1, 10)
		var onlineUsersColour = greaterThanSwitch(uonline, 1, 5)

		totonline, totunit := common.ConvertFriendlyUnit(totonline)
		uonline, uunit := common.ConvertFriendlyUnit(uonline)
		gonline, gunit := common.ConvertFriendlyUnit(gonline)

		addElement(common.GridElement{"dash-totonline", strconv.Itoa(totonline) + totunit + " online", 3, "grid_stat " + onlineColour, "", "", "The number of people who are currently online"})
		addElement(common.GridElement{"dash-gonline", strconv.Itoa(gonline) + gunit + " guests online", 4, "grid_stat " + onlineGuestsColour, "", "", "The number of guests who are currently online"})
		addElement(common.GridElement{"dash-uonline", strconv.Itoa(uonline) + uunit + " users online", 5, "grid_stat " + onlineUsersColour, "", "", "The number of logged-in users who are currently online"})
		addElement(common.GridElement{"dash-reqs", strconv.Itoa(reqCount) + " reqs / second", 7, "grid_stat grid_end_group " + topicColour, "", "", "The number of requests over the last 24 hours"})
	}

	addElement(common.GridElement{"dash-postsperday", strconv.Itoa(postCount) + " posts / " + postInterval, 6, "grid_stat " + postColour, "", "", "The number of new posts over the last 24 hours"})
	addElement(common.GridElement{"dash-topicsperday", strconv.Itoa(topicCount) + " topics / " + topicInterval, 7, "grid_stat " + topicColour, "", "", "The number of new topics over the last 24 hours"})
	addElement(common.GridElement{"dash-totonlineperday", "20 online / day", 8, "grid_stat stat_disabled", "", "", "Coming Soon!" /*, "The people online over the last 24 hours"*/})

	addElement(common.GridElement{"dash-searches", "8 searches / week", 9, "grid_stat stat_disabled", "", "", "Coming Soon!" /*"The number of searches over the last 7 days"*/})
	addElement(common.GridElement{"dash-newusers", strconv.Itoa(newUserCount) + " new users / " + newUserInterval, 10, "grid_stat", "", "", "The number of new users over the last 7 days"})
	addElement(common.GridElement{"dash-reports", strconv.Itoa(reportCount) + " reports / " + reportInterval, 11, "grid_stat", "", "", "The number of reports over the last 7 days"})

	if false {
		addElement(common.GridElement{"dash-minperuser", "2 minutes / user / week", 12, "grid_stat stat_disabled", "", "", "Coming Soon!" /*"The average number of number of minutes spent by each active user over the last 7 days"*/})
		addElement(common.GridElement{"dash-visitorsperweek", "2 visitors / week", 13, "grid_stat stat_disabled", "", "", "Coming Soon!" /*"The number of unique visitors we've had over the last 7 days"*/})
		addElement(common.GridElement{"dash-postsperuser", "5 posts / user / week", 14, "grid_stat stat_disabled", "", "", "Coming Soon!" /*"The average number of posts made by each active user over the past week"*/})
	}

	pi := common.PanelDashboardPage{&common.BasePanelPage{header, stats, "dashboard", common.ReportForumID}, gridElements}
	return panelRenderTemplate("panel_dashboard", w, r, user, &pi)
}

func routePanelWordFilters(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	header, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditSettings {
		return common.NoPermissions(w, r, user)
	}
	header.Title = common.GetTitlePhrase("panel_word_filters")

	var filterList = common.WordFilterBox.Load().(common.WordFilterMap)
	pi := common.PanelPage{&common.BasePanelPage{header, stats, "word-filters", common.ReportForumID}, tList, filterList}
	return panelRenderTemplate("panel_word_filters", w, r, user, &pi)
}

func routePanelWordFiltersCreateSubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditSettings {
		return common.NoPermissions(w, r, user)
	}
	isJs := (r.PostFormValue("js") == "1")

	// ? - We're not doing a full sanitise here, as it would be useful if admins were able to put down rules for replacing things with HTML, etc.
	find := strings.TrimSpace(r.PostFormValue("find"))
	if find == "" {
		return common.LocalErrorJSQ("You need to specify what word you want to match", w, r, user, isJs)
	}

	// Unlike with find, it's okay if we leave this blank, as this means that the admin wants to remove the word entirely with no replacement
	replacement := strings.TrimSpace(r.PostFormValue("replacement"))

	res, err := stmts.createWordFilter.Exec(find, replacement)
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	common.AddWordFilter(int(lastID), find, replacement)
	return panelSuccessRedirect("/panel/settings/word-filters/", w, r, isJs)
}

// TODO: Implement this as a non-JS fallback
func routePanelWordFiltersEdit(w http.ResponseWriter, r *http.Request, user common.User, wfid string) common.RouteError {
	header, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditSettings {
		return common.NoPermissions(w, r, user)
	}
	header.Title = common.GetTitlePhrase("panel_edit_word_filter")
	_ = wfid

	pi := common.PanelPage{&common.BasePanelPage{header, stats, "word-filters", common.ReportForumID}, tList, nil}
	return panelRenderTemplate("panel_word_filters_edit", w, r, user, &pi)
}

func routePanelWordFiltersEditSubmit(w http.ResponseWriter, r *http.Request, user common.User, wfid string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	// TODO: Either call it isJs or js rather than flip-flopping back and forth across the routes x.x
	isJs := (r.PostFormValue("isJs") == "1")
	if !user.Perms.EditSettings {
		return common.NoPermissionsJSQ(w, r, user, isJs)
	}

	id, err := strconv.Atoi(wfid)
	if err != nil {
		return common.LocalErrorJSQ("The word filter ID must be an integer.", w, r, user, isJs)
	}

	find := strings.TrimSpace(r.PostFormValue("find"))
	if find == "" {
		return common.LocalErrorJSQ("You need to specify what word you want to match", w, r, user, isJs)
	}

	// Unlike with find, it's okay if we leave this blank, as this means that the admin wants to remove the word entirely with no replacement
	replacement := strings.TrimSpace(r.PostFormValue("replacement"))

	_, err = stmts.updateWordFilter.Exec(find, replacement, id)
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	wordFilters := common.WordFilterBox.Load().(common.WordFilterMap)
	wordFilters[id] = common.WordFilter{ID: id, Find: find, Replacement: replacement}
	common.WordFilterBox.Store(wordFilters)

	http.Redirect(w, r, "/panel/settings/word-filters/", http.StatusSeeOther)
	return nil
}

func routePanelWordFiltersDeleteSubmit(w http.ResponseWriter, r *http.Request, user common.User, wfid string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	isJs := (r.PostFormValue("isJs") == "1")
	if !user.Perms.EditSettings {
		return common.NoPermissionsJSQ(w, r, user, isJs)
	}

	id, err := strconv.Atoi(wfid)
	if err != nil {
		return common.LocalErrorJSQ("The word filter ID must be an integer.", w, r, user, isJs)
	}

	_, err = stmts.deleteWordFilter.Exec(id)
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	wordFilters := common.WordFilterBox.Load().(common.WordFilterMap)
	delete(wordFilters, id)
	common.WordFilterBox.Store(wordFilters)

	http.Redirect(w, r, "/panel/settings/word-filters/", http.StatusSeeOther)
	return nil
}

func routePanelGroups(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	header, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	header.Title = common.GetTitlePhrase("panel_groups")

	page, _ := strconv.Atoi(r.FormValue("page"))
	perPage := 9
	offset, page, lastPage := common.PageOffset(stats.Groups, page, perPage)

	// Skip the 'Unknown' group
	offset++

	var count int
	var groupList []common.GroupAdmin
	groups, _ := common.Groups.GetRange(offset, 0)
	for _, group := range groups {
		if count == perPage {
			break
		}

		var rank string
		var rankClass string
		var canEdit bool
		var canDelete = false

		// TODO: Use a switch for this
		// TODO: Localise this
		if group.IsAdmin {
			rank = "Admin"
			rankClass = "admin"
		} else if group.IsMod {
			rank = "Mod"
			rankClass = "mod"
		} else if group.IsBanned {
			rank = "Banned"
			rankClass = "banned"
		} else if group.ID == 6 {
			rank = "Guest"
			rankClass = "guest"
		} else {
			rank = "Member"
			rankClass = "member"
		}

		canEdit = user.Perms.EditGroup && (!group.IsAdmin || user.Perms.EditGroupAdmin) && (!group.IsMod || user.Perms.EditGroupSuperMod)
		groupList = append(groupList, common.GroupAdmin{group.ID, group.Name, rank, rankClass, canEdit, canDelete})
		count++
	}
	//log.Printf("groupList: %+v\n", groupList)

	pageList := common.Paginate(stats.Groups, perPage, 5)
	pi := common.PanelGroupPage{&common.BasePanelPage{header, stats, "groups", common.ReportForumID}, groupList, common.Paginator{pageList, page, lastPage}}
	return panelRenderTemplate("panel_groups", w, r, user, &pi)
}

func routePanelGroupsEdit(w http.ResponseWriter, r *http.Request, user common.User, sgid string) common.RouteError {
	header, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditGroup {
		return common.NoPermissions(w, r, user)
	}
	header.Title = common.GetTitlePhrase("panel_edit_group")

	gid, err := strconv.Atoi(sgid)
	if err != nil {
		return common.LocalError("You need to provide a whole number for the group ID", w, r, user)
	}

	group, err := common.Groups.Get(gid)
	if err == ErrNoRows {
		//log.Print("aaaaa monsters")
		return common.NotFound(w, r, header)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	if group.IsAdmin && !user.Perms.EditGroupAdmin {
		return common.LocalError("You need the EditGroupAdmin permission to edit an admin group.", w, r, user)
	}
	if group.IsMod && !user.Perms.EditGroupSuperMod {
		return common.LocalError("You need the EditGroupSuperMod permission to edit a super-mod group.", w, r, user)
	}

	var rank string
	switch {
	case group.IsAdmin:
		rank = "Admin"
	case group.IsMod:
		rank = "Mod"
	case group.IsBanned:
		rank = "Banned"
	case group.ID == 6:
		rank = "Guest"
	default:
		rank = "Member"
	}

	disableRank := !user.Perms.EditGroupGlobalPerms || (group.ID == 6)

	pi := common.PanelEditGroupPage{&common.BasePanelPage{header, stats, "groups", common.ReportForumID}, group.ID, group.Name, group.Tag, rank, disableRank}
	if common.RunPreRenderHook("pre_render_panel_edit_group", w, r, &user, &pi) {
		return nil
	}
	err = common.Templates.ExecuteTemplate(w, "panel_group_edit.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

func routePanelGroupsEditPerms(w http.ResponseWriter, r *http.Request, user common.User, sgid string) common.RouteError {
	header, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditGroup {
		return common.NoPermissions(w, r, user)
	}
	header.Title = common.GetTitlePhrase("panel_edit_group")

	gid, err := strconv.Atoi(sgid)
	if err != nil {
		return common.LocalError("The Group ID is not a valid integer.", w, r, user)
	}

	group, err := common.Groups.Get(gid)
	if err == ErrNoRows {
		//log.Print("aaaaa monsters")
		return common.NotFound(w, r, header)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	if group.IsAdmin && !user.Perms.EditGroupAdmin {
		return common.LocalError("You need the EditGroupAdmin permission to edit an admin group.", w, r, user)
	}
	if group.IsMod && !user.Perms.EditGroupSuperMod {
		return common.LocalError("You need the EditGroupSuperMod permission to edit a super-mod group.", w, r, user)
	}

	// TODO: Load the phrases in bulk for efficiency?
	var localPerms []common.NameLangToggle

	var addLocalPerm = func(permStr string, perm bool) {
		localPerms = append(localPerms, common.NameLangToggle{permStr, common.GetLocalPermPhrase(permStr), perm})
	}

	addLocalPerm("ViewTopic", group.Perms.ViewTopic)
	addLocalPerm("LikeItem", group.Perms.LikeItem)
	addLocalPerm("CreateTopic", group.Perms.CreateTopic)
	//<--
	addLocalPerm("EditTopic", group.Perms.EditTopic)
	addLocalPerm("DeleteTopic", group.Perms.DeleteTopic)
	addLocalPerm("CreateReply", group.Perms.CreateReply)
	addLocalPerm("EditReply", group.Perms.EditReply)
	addLocalPerm("DeleteReply", group.Perms.DeleteReply)
	addLocalPerm("PinTopic", group.Perms.PinTopic)
	addLocalPerm("CloseTopic", group.Perms.CloseTopic)
	addLocalPerm("MoveTopic", group.Perms.MoveTopic)

	var globalPerms []common.NameLangToggle
	var addGlobalPerm = func(permStr string, perm bool) {
		globalPerms = append(globalPerms, common.NameLangToggle{permStr, common.GetGlobalPermPhrase(permStr), perm})
	}

	addGlobalPerm("BanUsers", group.Perms.BanUsers)
	addGlobalPerm("ActivateUsers", group.Perms.ActivateUsers)
	addGlobalPerm("EditUser", group.Perms.EditUser)
	addGlobalPerm("EditUserEmail", group.Perms.EditUserEmail)
	addGlobalPerm("EditUserPassword", group.Perms.EditUserPassword)
	addGlobalPerm("EditUserGroup", group.Perms.EditUserGroup)
	addGlobalPerm("EditUserGroupSuperMod", group.Perms.EditUserGroupSuperMod)
	addGlobalPerm("EditUserGroupAdmin", group.Perms.EditUserGroupAdmin)
	addGlobalPerm("EditGroup", group.Perms.EditGroup)
	addGlobalPerm("EditGroupLocalPerms", group.Perms.EditGroupLocalPerms)
	addGlobalPerm("EditGroupGlobalPerms", group.Perms.EditGroupGlobalPerms)
	addGlobalPerm("EditGroupSuperMod", group.Perms.EditGroupSuperMod)
	addGlobalPerm("EditGroupAdmin", group.Perms.EditGroupAdmin)
	addGlobalPerm("ManageForums", group.Perms.ManageForums)
	addGlobalPerm("EditSettings", group.Perms.EditSettings)
	addGlobalPerm("ManageThemes", group.Perms.ManageThemes)
	addGlobalPerm("ManagePlugins", group.Perms.ManagePlugins)
	addGlobalPerm("ViewAdminLogs", group.Perms.ViewAdminLogs)
	addGlobalPerm("ViewIPs", group.Perms.ViewIPs)
	addGlobalPerm("UploadFiles", group.Perms.UploadFiles)

	pi := common.PanelEditGroupPermsPage{&common.BasePanelPage{header, stats, "groups", common.ReportForumID}, group.ID, group.Name, localPerms, globalPerms}
	if common.RunPreRenderHook("pre_render_panel_edit_group_perms", w, r, &user, &pi) {
		return nil
	}
	err = common.Templates.ExecuteTemplate(w, "panel_group_edit_perms.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

func routePanelGroupsEditSubmit(w http.ResponseWriter, r *http.Request, user common.User, sgid string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditGroup {
		return common.NoPermissions(w, r, user)
	}

	gid, err := strconv.Atoi(sgid)
	if err != nil {
		return common.LocalError("You need to provide a whole number for the group ID", w, r, user)
	}

	group, err := common.Groups.Get(gid)
	if err == ErrNoRows {
		//log.Print("aaaaa monsters")
		return common.NotFound(w, r, nil)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	if group.IsAdmin && !user.Perms.EditGroupAdmin {
		return common.LocalError("You need the EditGroupAdmin permission to edit an admin group.", w, r, user)
	}
	if group.IsMod && !user.Perms.EditGroupSuperMod {
		return common.LocalError("You need the EditGroupSuperMod permission to edit a super-mod group.", w, r, user)
	}

	gname := r.FormValue("group-name")
	if gname == "" {
		return common.LocalError("The group name can't be left blank.", w, r, user)
	}
	gtag := r.FormValue("group-tag")
	rank := r.FormValue("group-type")

	var originalRank string
	// TODO: Use a switch for this
	if group.IsAdmin {
		originalRank = "Admin"
	} else if group.IsMod {
		originalRank = "Mod"
	} else if group.IsBanned {
		originalRank = "Banned"
	} else if group.ID == 6 {
		originalRank = "Guest"
	} else {
		originalRank = "Member"
	}

	if rank != originalRank {
		if !user.Perms.EditGroupGlobalPerms {
			return common.LocalError("You need the EditGroupGlobalPerms permission to change the group type.", w, r, user)
		}

		switch rank {
		case "Admin":
			if !user.Perms.EditGroupAdmin {
				return common.LocalError("You need the EditGroupAdmin permission to designate this group as an admin group.", w, r, user)
			}
			err = group.ChangeRank(true, true, false)
		case "Mod":
			if !user.Perms.EditGroupSuperMod {
				return common.LocalError("You need the EditGroupSuperMod permission to designate this group as a super-mod group.", w, r, user)
			}
			err = group.ChangeRank(false, true, false)
		case "Banned":
			err = group.ChangeRank(false, false, true)
		case "Guest":
			return common.LocalError("You can't designate a group as a guest group.", w, r, user)
		case "Member":
			err = group.ChangeRank(false, false, false)
		default:
			return common.LocalError("Invalid group type.", w, r, user)
		}
		if err != nil {
			return common.InternalError(err, w, r)
		}
	}

	// TODO: Move this to *Group
	_, err = stmts.updateGroup.Exec(gname, gtag, gid)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	common.Groups.Reload(gid)

	http.Redirect(w, r, "/panel/groups/edit/"+strconv.Itoa(gid), http.StatusSeeOther)
	return nil
}

func routePanelGroupsEditPermsSubmit(w http.ResponseWriter, r *http.Request, user common.User, sgid string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditGroup {
		return common.NoPermissions(w, r, user)
	}

	gid, err := strconv.Atoi(sgid)
	if err != nil {
		return common.LocalError("The Group ID is not a valid integer.", w, r, user)
	}

	group, err := common.Groups.Get(gid)
	if err == ErrNoRows {
		//log.Print("aaaaa monsters o.o")
		return common.NotFound(w, r, nil)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	if group.IsAdmin && !user.Perms.EditGroupAdmin {
		return common.LocalError("You need the EditGroupAdmin permission to edit an admin group.", w, r, user)
	}
	if group.IsMod && !user.Perms.EditGroupSuperMod {
		return common.LocalError("You need the EditGroupSuperMod permission to edit a super-mod group.", w, r, user)
	}

	var pmap = make(map[string]bool)
	if user.Perms.EditGroupLocalPerms {
		for _, perm := range common.LocalPermList {
			pvalue := r.PostFormValue("group-perm-" + perm)
			pmap[perm] = (pvalue == "1")
		}
	}

	if user.Perms.EditGroupGlobalPerms {
		for _, perm := range common.GlobalPermList {
			pvalue := r.PostFormValue("group-perm-" + perm)
			pmap[perm] = (pvalue == "1")
		}
	}

	// TODO: Abstract this
	pjson, err := json.Marshal(pmap)
	if err != nil {
		return common.LocalError("Unable to marshal the data", w, r, user)
	}
	_, err = stmts.updateGroupPerms.Exec(pjson, gid)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	err = common.RebuildGroupPermissions(gid)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/panel/groups/edit/perms/"+strconv.Itoa(gid), http.StatusSeeOther)
	return nil
}

func routePanelGroupsCreateSubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditGroup {
		return common.NoPermissions(w, r, user)
	}

	groupName := r.PostFormValue("group-name")
	if groupName == "" {
		return common.LocalError("You need a name for this group!", w, r, user)
	}
	groupTag := r.PostFormValue("group-tag")

	var isAdmin, isMod, isBanned bool
	if user.Perms.EditGroupGlobalPerms {
		groupType := r.PostFormValue("group-type")
		if groupType == "Admin" {
			if !user.Perms.EditGroupAdmin {
				return common.LocalError("You need the EditGroupAdmin permission to create admin groups", w, r, user)
			}
			isAdmin = true
			isMod = true
		} else if groupType == "Mod" {
			if !user.Perms.EditGroupSuperMod {
				return common.LocalError("You need the EditGroupSuperMod permission to create admin groups", w, r, user)
			}
			isMod = true
		} else if groupType == "Banned" {
			isBanned = true
		}
	}

	gid, err := common.Groups.Create(groupName, groupTag, isAdmin, isMod, isBanned)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	http.Redirect(w, r, "/panel/groups/edit/"+strconv.Itoa(gid), http.StatusSeeOther)
	return nil
}

func routePanelThemes(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	header, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageThemes {
		return common.NoPermissions(w, r, user)
	}
	header.Title = common.GetTitlePhrase("panel_themes")

	var pThemeList, vThemeList []*common.Theme
	for _, theme := range common.Themes {
		if theme.HideFromThemes {
			continue
		}
		if theme.ForkOf == "" {
			pThemeList = append(pThemeList, theme)
		} else {
			vThemeList = append(vThemeList, theme)
		}

	}

	pi := common.PanelThemesPage{&common.BasePanelPage{header, stats, "themes", common.ReportForumID}, pThemeList, vThemeList}
	return panelRenderTemplate("panel_themes", w, r, user, &pi)
}

func routePanelThemesSetDefault(w http.ResponseWriter, r *http.Request, user common.User, uname string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageThemes {
		return common.NoPermissions(w, r, user)
	}

	theme, ok := common.Themes[uname]
	if !ok {
		return common.LocalError("The theme isn't registered in the system", w, r, user)
	}
	if theme.Disabled {
		return common.LocalError("You must not enable this theme", w, r, user)
	}

	var isDefault bool
	err := stmts.isThemeDefault.QueryRow(uname).Scan(&isDefault)
	if err != nil && err != ErrNoRows {
		return common.InternalError(err, w, r)
	}

	hasTheme := err != ErrNoRows
	if hasTheme {
		if isDefault {
			return common.LocalError("The theme is already active", w, r, user)
		}
		_, err = stmts.updateTheme.Exec(1, uname)
	} else {
		_, err = stmts.addTheme.Exec(uname, 1)
	}
	if err != nil {
		return common.InternalError(err, w, r)
	}

	// TODO: Make this less racey
	// TODO: Move this to common
	common.ChangeDefaultThemeMutex.Lock()
	defaultTheme := common.DefaultThemeBox.Load().(string)
	_, err = stmts.updateTheme.Exec(0, defaultTheme)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	log.Printf("Setting theme '%s' as the default theme", theme.Name)
	theme.Active = true
	common.Themes[uname] = theme

	dTheme, ok := common.Themes[defaultTheme]
	if !ok {
		return common.InternalError(errors.New("The default theme is missing"), w, r)
	}
	dTheme.Active = false
	common.Themes[defaultTheme] = dTheme

	common.DefaultThemeBox.Store(uname)
	common.ResetTemplateOverrides()
	theme.MapTemplates()
	common.ChangeDefaultThemeMutex.Unlock()

	http.Redirect(w, r, "/panel/themes/", http.StatusSeeOther)
	return nil
}

func routePanelThemesMenus(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	header, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageThemes {
		return common.NoPermissions(w, r, user)
	}
	header.Title = common.GetTitlePhrase("panel_themes_menus")

	var menuList []common.PanelMenuListItem
	for mid, list := range common.Menus.GetAllMap() {
		var name = ""
		if mid == 1 {
			name = common.GetTmplPhrase("panel_themes_menus_main")
		}
		menuList = append(menuList, common.PanelMenuListItem{
			Name:      name,
			ID:        mid,
			ItemCount: len(list.List),
		})
	}

	pi := common.PanelMenuListPage{&common.BasePanelPage{header, stats, "themes", common.ReportForumID}, menuList}
	return panelRenderTemplate("panel_themes_menus", w, r, user, &pi)
}

func routePanelThemesMenusEdit(w http.ResponseWriter, r *http.Request, user common.User, smid string) common.RouteError {
	header, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageThemes {
		return common.NoPermissions(w, r, user)
	}
	// TODO: Something like Menu #1 for the title?
	header.Title = common.GetTitlePhrase("panel_themes_menus_edit")
	header.AddScript("Sortable-1.4.0/Sortable.min.js")

	mid, err := strconv.Atoi(smid)
	if err != nil {
		return common.LocalError("Invalid integer", w, r, user)
	}

	menuHold, err := common.Menus.Get(mid)
	if err == ErrNoRows {
		return common.NotFound(w, r, header)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	var menuList []common.MenuItem
	for _, item := range menuHold.List {
		var menuTmpls = map[string]common.MenuTmpl{
			item.TmplName: menuHold.Parse(item.Name, []byte("{{.Name}}")),
		}
		var renderBuffer [][]byte
		var variableIndices []int
		renderBuffer, _ = menuHold.ScanItem(menuTmpls, item, renderBuffer, variableIndices)

		var out string
		for _, renderItem := range renderBuffer {
			out += string(renderItem)
		}
		item.Name = out
		if item.Name == "" {
			item.Name = "???"
		}
		menuList = append(menuList, item)
	}

	pi := common.PanelMenuPage{&common.BasePanelPage{header, stats, "themes", common.ReportForumID}, mid, menuList}
	return panelRenderTemplate("panel_themes_menus_items", w, r, user, &pi)
}

func routePanelThemesMenuItemEdit(w http.ResponseWriter, r *http.Request, user common.User, sitemID string) common.RouteError {
	header, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageThemes {
		return common.NoPermissions(w, r, user)
	}
	// TODO: Something like Menu #1 for the title?
	header.Title = common.GetTitlePhrase("panel_themes_menus_edit")

	itemID, err := strconv.Atoi(sitemID)
	if err != nil {
		return common.LocalError("Invalid integer", w, r, user)
	}

	menuItem, err := common.Menus.ItemStore().Get(itemID)
	if err == ErrNoRows {
		return common.NotFound(w, r, header)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	pi := common.PanelMenuItemPage{&common.BasePanelPage{header, stats, "themes", common.ReportForumID}, menuItem}
	return panelRenderTemplate("panel_themes_menus_item_edit", w, r, user, &pi)
}

func routePanelThemesMenuItemSetters(r *http.Request, menuItem common.MenuItem) common.MenuItem {
	var getItem = func(name string) string {
		return common.SanitiseSingleLine(r.PostFormValue("item-" + name))
	}
	menuItem.Name = getItem("name")
	menuItem.HTMLID = getItem("htmlid")
	menuItem.CSSClass = getItem("cssclass")
	menuItem.Position = getItem("position")
	if menuItem.Position != "left" && menuItem.Position != "right" {
		menuItem.Position = "left"
	}
	menuItem.Path = getItem("path")
	menuItem.Aria = getItem("aria")
	menuItem.Tooltip = getItem("tooltip")
	menuItem.TmplName = getItem("tmplname")

	switch getItem("permissions") {
	case "everyone":
		menuItem.GuestOnly = false
		menuItem.MemberOnly = false
		menuItem.SuperModOnly = false
		menuItem.AdminOnly = false
	case "guest-only":
		menuItem.GuestOnly = true
		menuItem.MemberOnly = false
		menuItem.SuperModOnly = false
		menuItem.AdminOnly = false
	case "member-only":
		menuItem.GuestOnly = false
		menuItem.MemberOnly = true
		menuItem.SuperModOnly = false
		menuItem.AdminOnly = false
	case "supermod-only":
		menuItem.GuestOnly = false
		menuItem.MemberOnly = true
		menuItem.SuperModOnly = true
		menuItem.AdminOnly = false
	case "admin-only":
		menuItem.GuestOnly = false
		menuItem.MemberOnly = true
		menuItem.SuperModOnly = true
		menuItem.AdminOnly = true
	}
	return menuItem
}

func routePanelThemesMenuItemEditSubmit(w http.ResponseWriter, r *http.Request, user common.User, sitemID string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	isJs := (r.PostFormValue("js") == "1")
	if !user.Perms.ManageThemes {
		return common.NoPermissionsJSQ(w, r, user, isJs)
	}

	itemID, err := strconv.Atoi(sitemID)
	if err != nil {
		return common.LocalErrorJSQ("Invalid integer", w, r, user, isJs)
	}

	menuItem, err := common.Menus.ItemStore().Get(itemID)
	if err == ErrNoRows {
		return common.LocalErrorJSQ("This item doesn't exist.", w, r, user, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}
	//menuItem = menuItem.Copy() // If we switch this for a pointer, we might need this as a scratchpad
	menuItem = routePanelThemesMenuItemSetters(r, menuItem)

	err = menuItem.Commit()
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}
	return panelSuccessRedirect("/panel/themes/menus/item/edit/"+strconv.Itoa(itemID), w, r, isJs)
}

func routePanelThemesMenuItemCreateSubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	isJs := (r.PostFormValue("js") == "1")
	if !user.Perms.ManageThemes {
		return common.NoPermissionsJSQ(w, r, user, isJs)
	}
	smenuID := r.PostFormValue("mid")
	if smenuID == "" {
		return common.LocalErrorJSQ("No menuID provided", w, r, user, isJs)
	}
	menuID, err := strconv.Atoi(smenuID)
	if err != nil {
		return common.LocalErrorJSQ("Invalid integer", w, r, user, isJs)
	}

	menuItem := common.MenuItem{MenuID: menuID}
	menuItem = routePanelThemesMenuItemSetters(r, menuItem)
	itemID, err := menuItem.Create()
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}
	return panelSuccessRedirect("/panel/themes/menus/item/edit/"+strconv.Itoa(itemID), w, r, isJs)
}

func routePanelThemesMenuItemDeleteSubmit(w http.ResponseWriter, r *http.Request, user common.User, sitemID string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	isJs := (r.PostFormValue("js") == "1")
	if !user.Perms.ManageThemes {
		return common.NoPermissionsJSQ(w, r, user, isJs)
	}

	itemID, err := strconv.Atoi(sitemID)
	if err != nil {
		return common.LocalErrorJSQ("Invalid integer", w, r, user, isJs)
	}
	menuItem, err := common.Menus.ItemStore().Get(itemID)
	if err == ErrNoRows {
		return common.LocalErrorJSQ("This item doesn't exist.", w, r, user, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}
	//menuItem = menuItem.Copy() // If we switch this for a pointer, we might need this as a scratchpad

	err = menuItem.Delete()
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}
	return panelSuccessRedirect("/panel/themes/menus/", w, r, isJs)
}

func routePanelThemesMenuItemOrderSubmit(w http.ResponseWriter, r *http.Request, user common.User, smid string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	isJs := (r.PostFormValue("js") == "1")
	if !user.Perms.ManageThemes {
		return common.NoPermissionsJSQ(w, r, user, isJs)
	}

	mid, err := strconv.Atoi(smid)
	if err != nil {
		return common.LocalErrorJSQ("Invalid integer", w, r, user, isJs)
	}
	menuHold, err := common.Menus.Get(mid)
	if err == ErrNoRows {
		return common.LocalErrorJSQ("Can't find menu", w, r, user, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	sitems := strings.TrimSuffix(strings.TrimPrefix(r.PostFormValue("items"), "{"), "}")
	fmt.Printf("sitems: %+v\n", sitems)

	var updateMap = make(map[int]int)
	for index, smiid := range strings.Split(sitems, ",") {
		miid, err := strconv.Atoi(smiid)
		if err != nil {
			return common.LocalErrorJSQ("Invalid integer in menu item list", w, r, user, isJs)
		}
		updateMap[miid] = index
	}
	menuHold.UpdateOrder(updateMap)

	return panelSuccessRedirect("/panel/themes/menus/edit/"+strconv.Itoa(mid), w, r, isJs)
}
