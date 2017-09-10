// +build !no_templategen

// Code generated by Gosora. More below:
/* This file was automatically generated by the software. Please don't edit it as your changes may be overwritten at any moment. */
package main
import "net/http"
import "strconv"

// nolint
func init() {
	template_forum_handle = template_forum
	//o_template_forum_handle = template_forum
	ctemplates = append(ctemplates,"forum")
	tmplPtrMap["forum"] = &template_forum_handle
	tmplPtrMap["o_forum"] = template_forum
}

// nolint
func template_forum(tmpl_forum_vars ForumPage, w http.ResponseWriter) {
w.Write(header_0)
w.Write([]byte(tmpl_forum_vars.Title))
w.Write(header_1)
w.Write([]byte(tmpl_forum_vars.Header.ThemeName))
w.Write(header_2)
if len(tmpl_forum_vars.Header.Stylesheets) != 0 {
for _, item := range tmpl_forum_vars.Header.Stylesheets {
w.Write(header_3)
w.Write([]byte(item))
w.Write(header_4)
}
}
w.Write(header_5)
if len(tmpl_forum_vars.Header.Scripts) != 0 {
for _, item := range tmpl_forum_vars.Header.Scripts {
w.Write(header_6)
w.Write([]byte(item))
w.Write(header_7)
}
}
w.Write(header_8)
w.Write([]byte(tmpl_forum_vars.CurrentUser.Session))
w.Write(header_9)
if !tmpl_forum_vars.CurrentUser.IsSuperMod {
w.Write(header_10)
}
w.Write(header_11)
w.Write(menu_0)
w.Write(menu_1)
w.Write([]byte(tmpl_forum_vars.Header.Site.Name))
w.Write(menu_2)
if tmpl_forum_vars.CurrentUser.Loggedin {
w.Write(menu_3)
w.Write([]byte(tmpl_forum_vars.CurrentUser.Link))
w.Write(menu_4)
w.Write([]byte(tmpl_forum_vars.CurrentUser.Session))
w.Write(menu_5)
} else {
w.Write(menu_6)
}
w.Write(menu_7)
w.Write(header_12)
if tmpl_forum_vars.Header.Widgets.RightSidebar != "" {
w.Write(header_13)
}
w.Write(header_14)
if len(tmpl_forum_vars.Header.NoticeList) != 0 {
for _, item := range tmpl_forum_vars.Header.NoticeList {
w.Write(header_15)
w.Write([]byte(item))
w.Write(header_16)
}
}
if tmpl_forum_vars.Page > 1 {
w.Write(forum_0)
w.Write([]byte(strconv.Itoa(tmpl_forum_vars.Forum.ID)))
w.Write(forum_1)
w.Write([]byte(strconv.Itoa(tmpl_forum_vars.Page - 1)))
w.Write(forum_2)
}
if tmpl_forum_vars.LastPage != tmpl_forum_vars.Page {
w.Write(forum_3)
w.Write([]byte(strconv.Itoa(tmpl_forum_vars.Forum.ID)))
w.Write(forum_4)
w.Write([]byte(strconv.Itoa(tmpl_forum_vars.Page + 1)))
w.Write(forum_5)
w.Write([]byte(strconv.Itoa(tmpl_forum_vars.Forum.ID)))
w.Write(forum_6)
w.Write([]byte(strconv.Itoa(tmpl_forum_vars.Page + 1)))
w.Write(forum_7)
}
w.Write(forum_8)
if tmpl_forum_vars.CurrentUser.ID != 0 {
w.Write(forum_9)
}
w.Write(forum_10)
w.Write([]byte(tmpl_forum_vars.Title))
w.Write(forum_11)
if tmpl_forum_vars.CurrentUser.ID != 0 {
if tmpl_forum_vars.CurrentUser.Perms.CreateTopic {
w.Write(forum_12)
w.Write([]byte(strconv.Itoa(tmpl_forum_vars.Forum.ID)))
w.Write(forum_13)
} else {
w.Write(forum_14)
}
w.Write(forum_15)
}
w.Write(forum_16)
if len(tmpl_forum_vars.ItemList) != 0 {
for _, item := range tmpl_forum_vars.ItemList {
w.Write(forum_17)
if item.Sticky {
w.Write(forum_18)
} else {
if item.IsClosed {
w.Write(forum_19)
}
}
w.Write(forum_20)
if item.Creator.Avatar != "" {
w.Write(forum_21)
w.Write([]byte(item.Creator.Avatar))
w.Write(forum_22)
}
w.Write(forum_23)
w.Write([]byte(strconv.Itoa(item.PostCount)))
w.Write(forum_24)
w.Write([]byte(item.LastReplyAt))
w.Write(forum_25)
w.Write([]byte(item.Link))
w.Write(forum_26)
w.Write([]byte(item.Title))
w.Write(forum_27)
w.Write([]byte(item.Creator.Link))
w.Write(forum_28)
w.Write([]byte(item.Creator.Name))
w.Write(forum_29)
if item.IsClosed {
w.Write(forum_30)
}
if item.Sticky {
w.Write(forum_31)
}
w.Write(forum_32)
if item.LastUser.Avatar != "" {
w.Write(forum_33)
w.Write([]byte(item.LastUser.Avatar))
w.Write(forum_34)
}
w.Write(forum_35)
w.Write([]byte(item.LastUser.Link))
w.Write(forum_36)
w.Write([]byte(item.LastUser.Name))
w.Write(forum_37)
w.Write([]byte(item.LastReplyAt))
w.Write(forum_38)
}
} else {
w.Write(forum_39)
if tmpl_forum_vars.CurrentUser.Perms.CreateTopic {
w.Write(forum_40)
w.Write([]byte(strconv.Itoa(tmpl_forum_vars.Forum.ID)))
w.Write(forum_41)
}
w.Write(forum_42)
}
w.Write(forum_43)
w.Write(footer_0)
if len(tmpl_forum_vars.Header.Themes) != 0 {
for _, item := range tmpl_forum_vars.Header.Themes {
if !item.HideFromThemes {
w.Write(footer_1)
w.Write([]byte(item.Name))
w.Write(footer_2)
if tmpl_forum_vars.Header.ThemeName == item.Name {
w.Write(footer_3)
}
w.Write(footer_4)
w.Write([]byte(item.FriendlyName))
w.Write(footer_5)
}
}
}
w.Write(footer_6)
if tmpl_forum_vars.Header.Widgets.RightSidebar != "" {
w.Write(footer_7)
w.Write([]byte(string(tmpl_forum_vars.Header.Widgets.RightSidebar)))
w.Write(footer_8)
}
w.Write(footer_9)
}
