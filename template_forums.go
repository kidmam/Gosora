// Code generated by Gosora. More below:
/* This file was automatically generated by the software. Please don't edit it as your changes may be overwritten at any moment. */
// +build !no_templategen
package main
import "io"
import "strconv"

func init() {
	template_forums_handle = template_forums
	//o_template_forums_handle = template_forums
	ctemplates = append(ctemplates,"forums")
	tmpl_ptr_map["forums"] = &template_forums_handle
	tmpl_ptr_map["o_forums"] = template_forums
}

func template_forums(tmpl_forums_vars ForumsPage, w io.Writer) {
w.Write(header_0)
w.Write([]byte(tmpl_forums_vars.Title))
w.Write(header_1)
if len(tmpl_forums_vars.Header.Stylesheets) != 0 {
for _, item := range tmpl_forums_vars.Header.Stylesheets {
w.Write(header_2)
w.Write([]byte(item))
w.Write(header_3)
}
}
w.Write(header_4)
if len(tmpl_forums_vars.Header.Scripts) != 0 {
for _, item := range tmpl_forums_vars.Header.Scripts {
w.Write(header_5)
w.Write([]byte(item))
w.Write(header_6)
}
}
w.Write(header_7)
w.Write([]byte(tmpl_forums_vars.CurrentUser.Session))
w.Write(header_8)
w.Write(menu_0)
w.Write([]byte(tmpl_forums_vars.Header.Site.Name))
w.Write(menu_1)
if tmpl_forums_vars.CurrentUser.Loggedin {
w.Write(menu_2)
w.Write([]byte(tmpl_forums_vars.CurrentUser.Slug))
w.Write(menu_3)
w.Write([]byte(strconv.Itoa(tmpl_forums_vars.CurrentUser.ID)))
w.Write(menu_4)
if tmpl_forums_vars.CurrentUser.Is_Super_Mod {
w.Write(menu_5)
}
w.Write(menu_6)
w.Write([]byte(tmpl_forums_vars.CurrentUser.Session))
w.Write(menu_7)
} else {
w.Write(menu_8)
}
w.Write(menu_9)
w.Write(header_9)
if tmpl_forums_vars.Header.Widgets.RightSidebar != "" {
w.Write(header_10)
}
w.Write(header_11)
if len(tmpl_forums_vars.Header.NoticeList) != 0 {
for _, item := range tmpl_forums_vars.Header.NoticeList {
w.Write(header_12)
w.Write([]byte(item))
w.Write(header_13)
}
}
w.Write(forums_0)
if len(tmpl_forums_vars.ItemList) != 0 {
for _, item := range tmpl_forums_vars.ItemList {
w.Write(forums_1)
if item.Desc != "" || item.LastTopicTime != "" {
w.Write(forums_2)
}
w.Write(forums_3)
if item.Desc != "" {
w.Write(forums_4)
w.Write([]byte(item.Link))
w.Write(forums_5)
w.Write([]byte(item.Name))
w.Write(forums_6)
w.Write([]byte(item.Desc))
w.Write(forums_7)
} else {
w.Write(forums_8)
w.Write([]byte(item.Link))
w.Write(forums_9)
w.Write([]byte(item.Name))
w.Write(forums_10)
}
w.Write(forums_11)
w.Write([]byte(item.LastTopicSlug))
w.Write(forums_12)
w.Write([]byte(item.LastTopic))
w.Write(forums_13)
if item.LastTopicTime != "" {
w.Write(forums_14)
w.Write([]byte(item.LastTopicTime))
w.Write(forums_15)
}
w.Write(forums_16)
}
} else {
w.Write(forums_17)
}
w.Write(forums_18)
w.Write(footer_0)
if tmpl_forums_vars.Header.Widgets.RightSidebar != "" {
w.Write(footer_1)
w.Write([]byte(string(tmpl_forums_vars.Header.Widgets.RightSidebar)))
w.Write(footer_2)
}
w.Write(footer_3)
}
