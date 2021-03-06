package panel

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/phrases"
	"github.com/Azareal/Gosora/query_gen"
)

// TODO: Move this to another file, probably common/pages.go
type AnalyticsTimeRange struct {
	Quantity   int
	Unit       string
	Slices     int
	SliceWidth int
	Range      string
}

func analyticsTimeRange(rawTimeRange string) (timeRange AnalyticsTimeRange, err error) {
	timeRange.Quantity = 6
	timeRange.Unit = "hour"
	timeRange.Slices = 12
	timeRange.SliceWidth = 60 * 30
	timeRange.Range = "six-hours"

	switch rawTimeRange {
	// This might be pushing it, we might want to come up with a more efficient scheme for dealing with large timeframes like this
	case "one-year":
		timeRange.Quantity = 12
		timeRange.Unit = "month"
		timeRange.Slices = 12
		timeRange.SliceWidth = 60 * 60 * 24 * 30
		timeRange.Range = "one-year"
	case "three-months":
		timeRange.Quantity = 90
		timeRange.Unit = "day"
		timeRange.Slices = 30
		timeRange.SliceWidth = 60 * 60 * 24 * 3
		timeRange.Range = "three-months"
	case "one-month":
		timeRange.Quantity = 30
		timeRange.Unit = "day"
		timeRange.Slices = 30
		timeRange.SliceWidth = 60 * 60 * 24
		timeRange.Range = "one-month"
	case "one-week":
		timeRange.Quantity = 7
		timeRange.Unit = "day"
		timeRange.Slices = 14
		timeRange.SliceWidth = 60 * 60 * 12
		timeRange.Range = "one-week"
	case "two-days": // Two days is experimental
		timeRange.Quantity = 2
		timeRange.Unit = "day"
		timeRange.Slices = 24
		timeRange.SliceWidth = 60 * 60 * 2
		timeRange.Range = "two-days"
	case "one-day":
		timeRange.Quantity = 1
		timeRange.Unit = "day"
		timeRange.Slices = 24
		timeRange.SliceWidth = 60 * 60
		timeRange.Range = "one-day"
	case "twelve-hours":
		timeRange.Quantity = 12
		timeRange.Slices = 24
		timeRange.Range = "twelve-hours"
	case "six-hours", "":
	default:
		return timeRange, errors.New("Unknown time range")
	}
	return timeRange, nil
}

func analyticsTimeRangeToLabelList(timeRange AnalyticsTimeRange) (revLabelList []int64, labelList []int64, viewMap map[int64]int64) {
	viewMap = make(map[int64]int64)
	var currentTime = time.Now().Unix()
	for i := 1; i <= timeRange.Slices; i++ {
		var label = currentTime - int64(i*timeRange.SliceWidth)
		revLabelList = append(revLabelList, label)
		viewMap[label] = 0
	}
	for _, value := range revLabelList {
		labelList = append(labelList, value)
	}
	return revLabelList, labelList, viewMap
}

func analyticsRowsToViewMap(rows *sql.Rows, labelList []int64, viewMap map[int64]int64) (map[int64]int64, error) {
	defer rows.Close()
	for rows.Next() {
		var count int64
		var createdAt time.Time
		err := rows.Scan(&count, &createdAt)
		if err != nil {
			return viewMap, err
		}
		var unixCreatedAt = createdAt.Unix()
		// TODO: Bulk log this
		if common.Dev.SuperDebug {
			log.Print("count: ", count)
			log.Print("createdAt: ", createdAt)
			log.Print("unixCreatedAt: ", unixCreatedAt)
		}
		for _, value := range labelList {
			if unixCreatedAt > value {
				viewMap[value] += count
				break
			}
		}
	}
	return viewMap, rows.Err()
}

func PreAnalyticsDetail(w http.ResponseWriter, r *http.Request, user *common.User) (*common.BasePanelPage, common.RouteError) {
	basePage, ferr := buildBasePage(w, r, user, "analytics", "analytics")
	if ferr != nil {
		return nil, ferr
	}
	basePage.AddSheet("chartist/chartist.min.css")
	basePage.AddScript("chartist/chartist.min.js")
	basePage.AddScriptAsync("analytics.js")
	return basePage, nil
}

func AnalyticsViews(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	basePage, ferr := PreAnalyticsDetail(w, r, &user)
	if ferr != nil {
		return ferr
	}
	timeRange, err := analyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}
	revLabelList, labelList, viewMap := analyticsTimeRangeToLabelList(timeRange)

	common.DebugLog("in panel.AnalyticsViews")
	// TODO: Add some sort of analytics store / iterator?
	rows, err := qgen.NewAcc().Select("viewchunks").Columns("count, createdAt").Where("route = ''").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query()
	if err != nil && err != sql.ErrNoRows {
		return common.InternalError(err, w, r)
	}
	viewMap, err = analyticsRowsToViewMap(rows, labelList, viewMap)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	var viewList []int64
	var viewItems []common.PanelAnalyticsItem
	for _, value := range revLabelList {
		viewList = append(viewList, viewMap[value])
		viewItems = append(viewItems, common.PanelAnalyticsItem{Time: value, Count: viewMap[value]})
	}
	graph := common.PanelTimeGraph{Series: [][]int64{viewList}, Labels: labelList}
	common.DebugLogf("graph: %+v\n", graph)
	var ttime string
	if timeRange.Range == "six-hours" || timeRange.Range == "twelve-hours" || timeRange.Range == "one-day" {
		ttime = "time"
	}
	pi := common.PanelAnalyticsPage{basePage, graph, viewItems, timeRange.Range, timeRange.Unit, ttime}
	return renderTemplate("panel_analytics_views", w, r, basePage.Header, &pi)
}

func AnalyticsRouteViews(w http.ResponseWriter, r *http.Request, user common.User, route string) common.RouteError {
	basePage, ferr := PreAnalyticsDetail(w, r, &user)
	if ferr != nil {
		return ferr
	}
	timeRange, err := analyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}
	revLabelList, labelList, viewMap := analyticsTimeRangeToLabelList(timeRange)

	common.DebugLog("in panel.AnalyticsRouteViews")
	// TODO: Validate the route is valid
	rows, err := qgen.NewAcc().Select("viewchunks").Columns("count, createdAt").Where("route = ?").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query(route)
	if err != nil && err != sql.ErrNoRows {
		return common.InternalError(err, w, r)
	}
	viewMap, err = analyticsRowsToViewMap(rows, labelList, viewMap)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	var viewList []int64
	var viewItems []common.PanelAnalyticsItem
	for _, value := range revLabelList {
		viewList = append(viewList, viewMap[value])
		viewItems = append(viewItems, common.PanelAnalyticsItem{Time: value, Count: viewMap[value]})
	}
	graph := common.PanelTimeGraph{Series: [][]int64{viewList}, Labels: labelList}
	common.DebugLogf("graph: %+v\n", graph)

	pi := common.PanelAnalyticsRoutePage{basePage, common.SanitiseSingleLine(route), graph, viewItems, timeRange.Range}
	return renderTemplate("panel_analytics_route_views", w, r, basePage.Header, &pi)
}

func AnalyticsAgentViews(w http.ResponseWriter, r *http.Request, user common.User, agent string) common.RouteError {
	basePage, ferr := PreAnalyticsDetail(w, r, &user)
	if ferr != nil {
		return ferr
	}
	timeRange, err := analyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}
	revLabelList, labelList, viewMap := analyticsTimeRangeToLabelList(timeRange)
	// ? Only allow valid agents? The problem with this is that agents wind up getting renamed and it would take a migration to get them all up to snuff
	agent = common.SanitiseSingleLine(agent)

	common.DebugLog("in panel.AnalyticsAgentViews")
	// TODO: Verify the agent is valid
	rows, err := qgen.NewAcc().Select("viewchunks_agents").Columns("count, createdAt").Where("browser = ?").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query(agent)
	if err != nil && err != sql.ErrNoRows {
		return common.InternalError(err, w, r)
	}
	viewMap, err = analyticsRowsToViewMap(rows, labelList, viewMap)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	var viewList []int64
	for _, value := range revLabelList {
		viewList = append(viewList, viewMap[value])
	}
	graph := common.PanelTimeGraph{Series: [][]int64{viewList}, Labels: labelList}
	common.DebugLogf("graph: %+v\n", graph)

	friendlyAgent, ok := phrases.GetUserAgentPhrase(agent)
	if !ok {
		friendlyAgent = agent
	}

	pi := common.PanelAnalyticsAgentPage{basePage, agent, friendlyAgent, graph, timeRange.Range}
	return renderTemplate("panel_analytics_agent_views", w, r, basePage.Header, &pi)
}

func AnalyticsForumViews(w http.ResponseWriter, r *http.Request, user common.User, sfid string) common.RouteError {
	basePage, ferr := PreAnalyticsDetail(w, r, &user)
	if ferr != nil {
		return ferr
	}
	timeRange, err := analyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}
	revLabelList, labelList, viewMap := analyticsTimeRangeToLabelList(timeRange)

	fid, err := strconv.Atoi(sfid)
	if err != nil {
		return common.LocalError("Invalid integer", w, r, user)
	}

	common.DebugLog("in panel.AnalyticsForumViews")
	// TODO: Verify the agent is valid
	rows, err := qgen.NewAcc().Select("viewchunks_forums").Columns("count, createdAt").Where("forum = ?").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query(fid)
	if err != nil && err != sql.ErrNoRows {
		return common.InternalError(err, w, r)
	}
	viewMap, err = analyticsRowsToViewMap(rows, labelList, viewMap)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	var viewList []int64
	for _, value := range revLabelList {
		viewList = append(viewList, viewMap[value])
	}
	graph := common.PanelTimeGraph{Series: [][]int64{viewList}, Labels: labelList}
	common.DebugLogf("graph: %+v\n", graph)

	forum, err := common.Forums.Get(fid)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	pi := common.PanelAnalyticsAgentPage{basePage, sfid, forum.Name, graph, timeRange.Range}
	return renderTemplate("panel_analytics_forum_views", w, r, basePage.Header, &pi)
}

func AnalyticsSystemViews(w http.ResponseWriter, r *http.Request, user common.User, system string) common.RouteError {
	basePage, ferr := PreAnalyticsDetail(w, r, &user)
	if ferr != nil {
		return ferr
	}
	timeRange, err := analyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}
	revLabelList, labelList, viewMap := analyticsTimeRangeToLabelList(timeRange)
	system = common.SanitiseSingleLine(system)

	common.DebugLog("in panel.AnalyticsSystemViews")
	// TODO: Verify the OS name is valid
	rows, err := qgen.NewAcc().Select("viewchunks_systems").Columns("count, createdAt").Where("system = ?").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query(system)
	if err != nil && err != sql.ErrNoRows {
		return common.InternalError(err, w, r)
	}
	viewMap, err = analyticsRowsToViewMap(rows, labelList, viewMap)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	var viewList []int64
	for _, value := range revLabelList {
		viewList = append(viewList, viewMap[value])
	}
	graph := common.PanelTimeGraph{Series: [][]int64{viewList}, Labels: labelList}
	common.DebugLogf("graph: %+v\n", graph)

	friendlySystem, ok := phrases.GetOSPhrase(system)
	if !ok {
		friendlySystem = system
	}

	pi := common.PanelAnalyticsAgentPage{basePage, system, friendlySystem, graph, timeRange.Range}
	return renderTemplate("panel_analytics_system_views", w, r, basePage.Header, &pi)
}

func AnalyticsLanguageViews(w http.ResponseWriter, r *http.Request, user common.User, lang string) common.RouteError {
	basePage, ferr := PreAnalyticsDetail(w, r, &user)
	if ferr != nil {
		return ferr
	}
	timeRange, err := analyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}
	revLabelList, labelList, viewMap := analyticsTimeRangeToLabelList(timeRange)
	lang = common.SanitiseSingleLine(lang)

	common.DebugLog("in panel.AnalyticsLanguageViews")
	// TODO: Verify the language code is valid
	rows, err := qgen.NewAcc().Select("viewchunks_langs").Columns("count, createdAt").Where("lang = ?").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query(lang)
	if err != nil && err != sql.ErrNoRows {
		return common.InternalError(err, w, r)
	}

	viewMap, err = analyticsRowsToViewMap(rows, labelList, viewMap)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	var viewList []int64
	for _, value := range revLabelList {
		viewList = append(viewList, viewMap[value])
	}
	graph := common.PanelTimeGraph{Series: [][]int64{viewList}, Labels: labelList}
	common.DebugLogf("graph: %+v\n", graph)

	friendlyLang, ok := phrases.GetHumanLangPhrase(lang)
	if !ok {
		friendlyLang = lang
	}

	pi := common.PanelAnalyticsAgentPage{basePage, lang, friendlyLang, graph, timeRange.Range}
	return renderTemplate("panel_analytics_lang_views", w, r, basePage.Header, &pi)
}

func AnalyticsReferrerViews(w http.ResponseWriter, r *http.Request, user common.User, domain string) common.RouteError {
	basePage, ferr := PreAnalyticsDetail(w, r, &user)
	if ferr != nil {
		return ferr
	}
	timeRange, err := analyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}
	revLabelList, labelList, viewMap := analyticsTimeRangeToLabelList(timeRange)

	common.DebugLog("in panel.AnalyticsReferrerViews")
	// TODO: Verify the agent is valid
	rows, err := qgen.NewAcc().Select("viewchunks_referrers").Columns("count, createdAt").Where("domain = ?").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query(domain)
	if err != nil && err != sql.ErrNoRows {
		return common.InternalError(err, w, r)
	}
	viewMap, err = analyticsRowsToViewMap(rows, labelList, viewMap)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	var viewList []int64
	for _, value := range revLabelList {
		viewList = append(viewList, viewMap[value])
	}
	graph := common.PanelTimeGraph{Series: [][]int64{viewList}, Labels: labelList}
	common.DebugLogf("graph: %+v\n", graph)
	pi := common.PanelAnalyticsAgentPage{basePage, common.SanitiseSingleLine(domain), "", graph, timeRange.Range}
	return renderTemplate("panel_analytics_referrer_views", w, r, basePage.Header, &pi)
}

func AnalyticsTopics(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	basePage, ferr := PreAnalyticsDetail(w, r, &user)
	if ferr != nil {
		return ferr
	}
	timeRange, err := analyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}
	revLabelList, labelList, viewMap := analyticsTimeRangeToLabelList(timeRange)

	common.DebugLog("in panel.AnalyticsTopics")
	rows, err := qgen.NewAcc().Select("topicchunks").Columns("count, createdAt").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query()
	if err != nil && err != sql.ErrNoRows {
		return common.InternalError(err, w, r)
	}
	viewMap, err = analyticsRowsToViewMap(rows, labelList, viewMap)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	var viewList []int64
	var viewItems []common.PanelAnalyticsItem
	for _, value := range revLabelList {
		viewList = append(viewList, viewMap[value])
		viewItems = append(viewItems, common.PanelAnalyticsItem{Time: value, Count: viewMap[value]})
	}
	graph := common.PanelTimeGraph{Series: [][]int64{viewList}, Labels: labelList}
	common.DebugLogf("graph: %+v\n", graph)
	pi := common.PanelAnalyticsPage{basePage, graph, viewItems, timeRange.Range, timeRange.Unit, "time"}
	return renderTemplate("panel_analytics_topics", w, r, basePage.Header, &pi)
}

func AnalyticsPosts(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	basePage, ferr := PreAnalyticsDetail(w, r, &user)
	if ferr != nil {
		return ferr
	}
	timeRange, err := analyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}
	revLabelList, labelList, viewMap := analyticsTimeRangeToLabelList(timeRange)

	common.DebugLog("in panel.AnalyticsPosts")
	rows, err := qgen.NewAcc().Select("postchunks").Columns("count, createdAt").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query()
	if err != nil && err != sql.ErrNoRows {
		return common.InternalError(err, w, r)
	}
	viewMap, err = analyticsRowsToViewMap(rows, labelList, viewMap)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	var viewList []int64
	var viewItems []common.PanelAnalyticsItem
	for _, value := range revLabelList {
		viewList = append(viewList, viewMap[value])
		viewItems = append(viewItems, common.PanelAnalyticsItem{Time: value, Count: viewMap[value]})
	}
	graph := common.PanelTimeGraph{Series: [][]int64{viewList}, Labels: labelList}
	common.DebugLogf("graph: %+v\n", graph)
	pi := common.PanelAnalyticsPage{basePage, graph, viewItems, timeRange.Range, timeRange.Unit, "time"}
	return renderTemplate("panel_analytics_posts", w, r, basePage.Header, &pi)
}

func analyticsRowsToNameMap(rows *sql.Rows) (map[string]int, error) {
	nameMap := make(map[string]int)
	defer rows.Close()
	for rows.Next() {
		var count int
		var name string
		err := rows.Scan(&count, &name)
		if err != nil {
			return nameMap, err
		}
		// TODO: Bulk log this
		if common.Dev.SuperDebug {
			log.Print("count: ", count)
			log.Print("name: ", name)
		}
		nameMap[name] += count
	}
	return nameMap, rows.Err()
}

func analyticsRowsToDuoMap(rows *sql.Rows, labelList []int64, viewMap map[int64]int64) (map[string]map[int64]int64, map[string]int, error) {
	vMap := make(map[string]map[int64]int64)
	nameMap := make(map[string]int)
	defer rows.Close()
	for rows.Next() {
		var count int64
		var name string
		var createdAt time.Time
		err := rows.Scan(&count, &name, &createdAt)
		if err != nil {
			return vMap, nameMap, err
		}

		// TODO: Bulk log this
		var unixCreatedAt = createdAt.Unix()
		if common.Dev.SuperDebug {
			log.Print("count: ", count)
			log.Print("name: ", name)
			log.Print("createdAt: ", createdAt)
			log.Print("unixCreatedAt: ", unixCreatedAt)
		}
		vvMap, ok := vMap[name]
		if !ok {
			vvMap = make(map[int64]int64)
			for key, val := range viewMap {
				vvMap[key] = val
			}
			vMap[name] = vvMap
		}
		for _, value := range labelList {
			if unixCreatedAt > value {
				vvMap[value] += count
				break
			}
		}
		nameMap[name] += int(count)
	}
	return vMap, nameMap, rows.Err()
}

type OVItem struct {
	name    string
	count   int
	viewMap map[int64]int64
}

func analyticsVMapToOVList(vMap map[string]map[int64]int64) (ovList []OVItem) {
	// Order the map
	for name, viewMap := range vMap {
		var totcount int
		for _, count := range viewMap {
			totcount += int(count)
		}
		ovList = append(ovList, OVItem{name, totcount, viewMap})
	}

	// Use bubble sort for now as there shouldn't be too many items
	for i := 0; i < len(ovList)-1; i++ {
		for j := 0; j < len(ovList)-1; j++ {
			if ovList[j].count > ovList[j+1].count {
				temp := ovList[j]
				ovList[j] = ovList[j+1]
				ovList[j+1] = temp
			}
		}
	}

	// Invert the direction
	var tOVList []OVItem
	for i := len(ovList) - 1; i >= 0; i-- {
		tOVList = append(tOVList, ovList[i])
	}
	return tOVList
}

func AnalyticsForums(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	basePage, ferr := PreAnalyticsDetail(w, r, &user)
	if ferr != nil {
		return ferr
	}
	basePage.AddScript("chartist/chartist-plugin-legend.min.js")
	basePage.AddSheet("chartist/chartist-plugin-legend.css")

	timeRange, err := analyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}
	revLabelList, labelList, viewMap := analyticsTimeRangeToLabelList(timeRange)

	rows, err := qgen.NewAcc().Select("viewchunks_forums").Columns("count, forum, createdAt").Where("forum != ''").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query()
	if err != nil && err != sql.ErrNoRows {
		return common.InternalError(err, w, r)
	}
	vMap, forumMap, err := analyticsRowsToDuoMap(rows, labelList, viewMap)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	ovList := analyticsVMapToOVList(vMap)

	var vList [][]int64
	var legendList []string
	var i int
	for _, ovitem := range ovList {
		var viewList []int64
		for _, value := range revLabelList {
			viewList = append(viewList, ovitem.viewMap[value])
		}
		vList = append(vList, viewList)
		fid, err := strconv.Atoi(ovitem.name)
		if err != nil {
			return common.InternalError(err, w, r)
		}
		var lName string
		forum, err := common.Forums.Get(fid)
		if err == sql.ErrNoRows {
			// TODO: Localise this
			lName = "Deleted Forum"
		} else if err != nil {
			return common.InternalError(err, w, r)
		} else {
			lName = forum.Name
		}
		legendList = append(legendList, lName)
		if i >= 6 {
			break
		}
		i++
	}
	graph := common.PanelTimeGraph{Series: vList, Labels: labelList, Legends: legendList}
	common.DebugLogf("graph: %+v\n", graph)

	// TODO: Sort this slice
	var forumItems []common.PanelAnalyticsAgentsItem
	for sfid, count := range forumMap {
		fid, err := strconv.Atoi(sfid)
		if err != nil {
			return common.InternalError(err, w, r)
		}
		var lName string
		forum, err := common.Forums.Get(fid)
		if err == sql.ErrNoRows {
			// TODO: Localise this
			lName = "Deleted Forum"
		} else if err != nil {
			return common.InternalError(err, w, r)
		} else {
			lName = forum.Name
		}
		forumItems = append(forumItems, common.PanelAnalyticsAgentsItem{
			Agent:         sfid,
			FriendlyAgent: lName,
			Count:         count,
		})
	}

	pi := common.PanelAnalyticsDuoPage{basePage, forumItems, graph, timeRange.Range}
	return renderTemplate("panel_analytics_forums", w, r, basePage.Header, &pi)
}

func AnalyticsRoutes(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	basePage, ferr := PreAnalyticsDetail(w, r, &user)
	if ferr != nil {
		return ferr
	}
	basePage.AddScript("chartist/chartist-plugin-legend.min.js")
	basePage.AddSheet("chartist/chartist-plugin-legend.css")

	timeRange, err := analyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}
	revLabelList, labelList, viewMap := analyticsTimeRangeToLabelList(timeRange)

	rows, err := qgen.NewAcc().Select("viewchunks").Columns("count, route, createdAt").Where("route != ''").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query()
	if err != nil && err != sql.ErrNoRows {
		return common.InternalError(err, w, r)
	}
	vMap, routeMap, err := analyticsRowsToDuoMap(rows, labelList, viewMap)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	ovList := analyticsVMapToOVList(vMap)

	var vList [][]int64
	var legendList []string
	var i int
	for _, ovitem := range ovList {
		var viewList []int64
		for _, value := range revLabelList {
			viewList = append(viewList, ovitem.viewMap[value])
		}
		vList = append(vList, viewList)
		legendList = append(legendList, ovitem.name)
		if i >= 6 {
			break
		}
		i++
	}
	graph := common.PanelTimeGraph{Series: vList, Labels: labelList, Legends: legendList}
	common.DebugLogf("graph: %+v\n", graph)

	// TODO: Sort this slice
	var routeItems []common.PanelAnalyticsRoutesItem
	for route, count := range routeMap {
		routeItems = append(routeItems, common.PanelAnalyticsRoutesItem{
			Route: route,
			Count: count,
		})
	}

	pi := common.PanelAnalyticsRoutesPage{basePage, routeItems, graph, timeRange.Range}
	return renderTemplate("panel_analytics_routes", w, r, basePage.Header, &pi)
}

// Trialling multi-series charts
func AnalyticsAgents(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	basePage, ferr := PreAnalyticsDetail(w, r, &user)
	if ferr != nil {
		return ferr
	}
	basePage.AddScript("chartist/chartist-plugin-legend.min.js")
	basePage.AddSheet("chartist/chartist-plugin-legend.css")

	timeRange, err := analyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}
	revLabelList, labelList, viewMap := analyticsTimeRangeToLabelList(timeRange)

	rows, err := qgen.NewAcc().Select("viewchunks_agents").Columns("count, browser, createdAt").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query()
	if err != nil && err != sql.ErrNoRows {
		return common.InternalError(err, w, r)
	}
	vMap, agentMap, err := analyticsRowsToDuoMap(rows, labelList, viewMap)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	ovList := analyticsVMapToOVList(vMap)

	var vList [][]int64
	var legendList []string
	var i int
	for _, ovitem := range ovList {
		var viewList []int64
		for _, value := range revLabelList {
			viewList = append(viewList, ovitem.viewMap[value])
		}
		vList = append(vList, viewList)
		lName, ok := phrases.GetUserAgentPhrase(ovitem.name)
		if !ok {
			lName = ovitem.name
		}
		legendList = append(legendList, lName)
		if i >= 6 {
			break
		}
		i++
	}
	graph := common.PanelTimeGraph{Series: vList, Labels: labelList, Legends: legendList}
	common.DebugLogf("graph: %+v\n", graph)

	// TODO: Sort this slice
	var agentItems []common.PanelAnalyticsAgentsItem
	for agent, count := range agentMap {
		aAgent, ok := phrases.GetUserAgentPhrase(agent)
		if !ok {
			aAgent = agent
		}
		agentItems = append(agentItems, common.PanelAnalyticsAgentsItem{
			Agent:         agent,
			FriendlyAgent: aAgent,
			Count:         count,
		})
	}

	pi := common.PanelAnalyticsDuoPage{basePage, agentItems, graph, timeRange.Range}
	return renderTemplate("panel_analytics_agents", w, r, basePage.Header, &pi)
}

func AnalyticsSystems(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	basePage, ferr := PreAnalyticsDetail(w, r, &user)
	if ferr != nil {
		return ferr
	}
	basePage.AddScript("chartist/chartist-plugin-legend.min.js")
	basePage.AddSheet("chartist/chartist-plugin-legend.css")

	timeRange, err := analyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}
	revLabelList, labelList, viewMap := analyticsTimeRangeToLabelList(timeRange)

	rows, err := qgen.NewAcc().Select("viewchunks_systems").Columns("count, system, createdAt").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query()
	if err != nil && err != sql.ErrNoRows {
		return common.InternalError(err, w, r)
	}
	vMap, osMap, err := analyticsRowsToDuoMap(rows, labelList, viewMap)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	ovList := analyticsVMapToOVList(vMap)

	var vList [][]int64
	var legendList []string
	var i int
	for _, ovitem := range ovList {
		var viewList []int64
		for _, value := range revLabelList {
			viewList = append(viewList, ovitem.viewMap[value])
		}
		vList = append(vList, viewList)
		lName, ok := phrases.GetOSPhrase(ovitem.name)
		if !ok {
			lName = ovitem.name
		}
		legendList = append(legendList, lName)
		if i >= 6 {
			break
		}
		i++
	}
	graph := common.PanelTimeGraph{Series: vList, Labels: labelList, Legends: legendList}
	common.DebugLogf("graph: %+v\n", graph)

	// TODO: Sort this slice
	var systemItems []common.PanelAnalyticsAgentsItem
	for system, count := range osMap {
		sSystem, ok := phrases.GetOSPhrase(system)
		if !ok {
			sSystem = system
		}
		systemItems = append(systemItems, common.PanelAnalyticsAgentsItem{
			Agent:         system,
			FriendlyAgent: sSystem,
			Count:         count,
		})
	}

	pi := common.PanelAnalyticsDuoPage{basePage, systemItems, graph, timeRange.Range}
	return renderTemplate("panel_analytics_systems", w, r, basePage.Header, &pi)
}

func AnalyticsLanguages(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	basePage, ferr := PreAnalyticsDetail(w, r, &user)
	if ferr != nil {
		return ferr
	}
	basePage.AddScript("chartist/chartist-plugin-legend.min.js")
	basePage.AddSheet("chartist/chartist-plugin-legend.css")

	timeRange, err := analyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}
	revLabelList, labelList, viewMap := analyticsTimeRangeToLabelList(timeRange)

	rows, err := qgen.NewAcc().Select("viewchunks_langs").Columns("count, lang, createdAt").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query()
	if err != nil && err != sql.ErrNoRows {
		return common.InternalError(err, w, r)
	}
	vMap, langMap, err := analyticsRowsToDuoMap(rows, labelList, viewMap)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	ovList := analyticsVMapToOVList(vMap)

	var vList [][]int64
	var legendList []string
	var i int
	for _, ovitem := range ovList {
		var viewList []int64
		for _, value := range revLabelList {
			viewList = append(viewList, ovitem.viewMap[value])
		}
		vList = append(vList, viewList)
		lName, ok := phrases.GetHumanLangPhrase(ovitem.name)
		if !ok {
			lName = ovitem.name
		}
		legendList = append(legendList, lName)
		if i >= 6 {
			break
		}
		i++
	}
	graph := common.PanelTimeGraph{Series: vList, Labels: labelList, Legends: legendList}
	common.DebugLogf("graph: %+v\n", graph)

	// TODO: Can we de-duplicate these analytics functions further?
	// TODO: Sort this slice
	var langItems []common.PanelAnalyticsAgentsItem
	for lang, count := range langMap {
		lLang, ok := phrases.GetHumanLangPhrase(lang)
		if !ok {
			lLang = lang
		}
		langItems = append(langItems, common.PanelAnalyticsAgentsItem{
			Agent:         lang,
			FriendlyAgent: lLang,
			Count:         count,
		})
	}

	pi := common.PanelAnalyticsDuoPage{basePage, langItems, graph, timeRange.Range}
	return renderTemplate("panel_analytics_langs", w, r, basePage.Header, &pi)
}

func AnalyticsReferrers(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	basePage, ferr := buildBasePage(w, r, &user, "analytics", "analytics")
	if ferr != nil {
		return ferr
	}
	timeRange, err := analyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}

	rows, err := qgen.NewAcc().Select("viewchunks_referrers").Columns("count, domain").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query()
	if err != nil && err != sql.ErrNoRows {
		return common.InternalError(err, w, r)
	}
	refMap, err := analyticsRowsToNameMap(rows)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	// TODO: Sort this slice
	var refItems []common.PanelAnalyticsAgentsItem
	for domain, count := range refMap {
		refItems = append(refItems, common.PanelAnalyticsAgentsItem{
			Agent: common.SanitiseSingleLine(domain),
			Count: count,
		})
	}

	pi := common.PanelAnalyticsAgentsPage{basePage, refItems, timeRange.Range}
	return renderTemplate("panel_analytics_referrers", w, r, basePage.Header, &pi)
}
