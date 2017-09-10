// +build !no_templategen

// Code generated by Gosora. More below:
/* This file was automatically generated by the software. Please don't edit it as your changes may be overwritten at any moment. */
package main
import "net/http"
import "strconv"

// nolint
func init() {
	template_topic_alt_handle = template_topic_alt
	//o_template_topic_alt_handle = template_topic_alt
	ctemplates = append(ctemplates,"topic_alt")
	tmplPtrMap["topic_alt"] = &template_topic_alt_handle
	tmplPtrMap["o_topic_alt"] = template_topic_alt
}

// nolint
func template_topic_alt(tmpl_topic_alt_vars TopicPage, w http.ResponseWriter) {
w.Write(header_0)
w.Write([]byte(tmpl_topic_alt_vars.Title))
w.Write(header_1)
w.Write([]byte(tmpl_topic_alt_vars.Header.ThemeName))
w.Write(header_2)
if len(tmpl_topic_alt_vars.Header.Stylesheets) != 0 {
for _, item := range tmpl_topic_alt_vars.Header.Stylesheets {
w.Write(header_3)
w.Write([]byte(item))
w.Write(header_4)
}
}
w.Write(header_5)
if len(tmpl_topic_alt_vars.Header.Scripts) != 0 {
for _, item := range tmpl_topic_alt_vars.Header.Scripts {
w.Write(header_6)
w.Write([]byte(item))
w.Write(header_7)
}
}
w.Write(header_8)
w.Write([]byte(tmpl_topic_alt_vars.CurrentUser.Session))
w.Write(header_9)
if !tmpl_topic_alt_vars.CurrentUser.IsSuperMod {
w.Write(header_10)
}
w.Write(header_11)
w.Write(menu_0)
w.Write(menu_1)
w.Write([]byte(tmpl_topic_alt_vars.Header.Site.Name))
w.Write(menu_2)
if tmpl_topic_alt_vars.CurrentUser.Loggedin {
w.Write(menu_3)
w.Write([]byte(tmpl_topic_alt_vars.CurrentUser.Link))
w.Write(menu_4)
w.Write([]byte(tmpl_topic_alt_vars.CurrentUser.Session))
w.Write(menu_5)
} else {
w.Write(menu_6)
}
w.Write(menu_7)
w.Write(header_12)
if tmpl_topic_alt_vars.Header.Widgets.RightSidebar != "" {
w.Write(header_13)
}
w.Write(header_14)
if len(tmpl_topic_alt_vars.Header.NoticeList) != 0 {
for _, item := range tmpl_topic_alt_vars.Header.NoticeList {
w.Write(header_15)
w.Write([]byte(item))
w.Write(header_16)
}
}
if tmpl_topic_alt_vars.Page > 1 {
w.Write(topic_alt_0)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Topic.ID)))
w.Write(topic_alt_1)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Page - 1)))
w.Write(topic_alt_2)
}
if tmpl_topic_alt_vars.LastPage != tmpl_topic_alt_vars.Page {
w.Write(topic_alt_3)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Topic.ID)))
w.Write(topic_alt_4)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Page + 1)))
w.Write(topic_alt_5)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Topic.ID)))
w.Write(topic_alt_6)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Page + 1)))
w.Write(topic_alt_7)
}
w.Write(topic_alt_8)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Topic.ID)))
w.Write(topic_alt_9)
if tmpl_topic_alt_vars.Topic.Sticky {
w.Write(topic_alt_10)
} else {
if tmpl_topic_alt_vars.Topic.IsClosed {
w.Write(topic_alt_11)
}
}
w.Write(topic_alt_12)
w.Write([]byte(tmpl_topic_alt_vars.Topic.Title))
w.Write(topic_alt_13)
if tmpl_topic_alt_vars.Topic.IsClosed {
w.Write(topic_alt_14)
}
if tmpl_topic_alt_vars.CurrentUser.Perms.EditTopic {
w.Write(topic_alt_15)
w.Write([]byte(tmpl_topic_alt_vars.Topic.Title))
w.Write(topic_alt_16)
if tmpl_topic_alt_vars.CurrentUser.Perms.CloseTopic {
w.Write(topic_alt_17)
}
w.Write(topic_alt_18)
}
w.Write(topic_alt_19)
w.Write([]byte(tmpl_topic_alt_vars.Topic.Avatar))
w.Write(topic_alt_20)
w.Write([]byte(tmpl_topic_alt_vars.Topic.UserLink))
w.Write(topic_alt_21)
w.Write([]byte(tmpl_topic_alt_vars.Topic.CreatedByName))
w.Write(topic_alt_22)
if tmpl_topic_alt_vars.Topic.Tag != "" {
w.Write(topic_alt_23)
w.Write([]byte(tmpl_topic_alt_vars.Topic.Tag))
w.Write(topic_alt_24)
} else {
w.Write(topic_alt_25)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Topic.Level)))
w.Write(topic_alt_26)
}
w.Write(topic_alt_27)
w.Write([]byte(tmpl_topic_alt_vars.Topic.Content))
w.Write(topic_alt_28)
w.Write([]byte(tmpl_topic_alt_vars.Topic.Content))
w.Write(topic_alt_29)
if tmpl_topic_alt_vars.CurrentUser.Loggedin {
if tmpl_topic_alt_vars.CurrentUser.Perms.LikeItem {
w.Write(topic_alt_30)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Topic.ID)))
w.Write(topic_alt_31)
}
if tmpl_topic_alt_vars.CurrentUser.Perms.EditTopic {
w.Write(topic_alt_32)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Topic.ID)))
w.Write(topic_alt_33)
}
if tmpl_topic_alt_vars.CurrentUser.Perms.DeleteTopic {
w.Write(topic_alt_34)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Topic.ID)))
w.Write(topic_alt_35)
}
if tmpl_topic_alt_vars.CurrentUser.Perms.PinTopic {
if tmpl_topic_alt_vars.Topic.Sticky {
w.Write(topic_alt_36)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Topic.ID)))
w.Write(topic_alt_37)
} else {
w.Write(topic_alt_38)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Topic.ID)))
w.Write(topic_alt_39)
}
}
w.Write(topic_alt_40)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Topic.ID)))
w.Write(topic_alt_41)
w.Write([]byte(tmpl_topic_alt_vars.CurrentUser.Session))
w.Write(topic_alt_42)
if tmpl_topic_alt_vars.CurrentUser.Perms.ViewIPs {
w.Write(topic_alt_43)
w.Write([]byte(tmpl_topic_alt_vars.Topic.IPAddress))
w.Write(topic_alt_44)
}
}
w.Write(topic_alt_45)
w.Write([]byte(tmpl_topic_alt_vars.Topic.CreatedAt))
w.Write(topic_alt_46)
if tmpl_topic_alt_vars.Topic.LikeCount > 0 {
w.Write(topic_alt_47)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Topic.LikeCount)))
w.Write(topic_alt_48)
}
w.Write(topic_alt_49)
if len(tmpl_topic_alt_vars.ItemList) != 0 {
for _, item := range tmpl_topic_alt_vars.ItemList {
w.Write(topic_alt_50)
if item.ActionType != "" {
w.Write(topic_alt_51)
}
w.Write(topic_alt_52)
w.Write([]byte(item.Avatar))
w.Write(topic_alt_53)
w.Write([]byte(item.UserLink))
w.Write(topic_alt_54)
w.Write([]byte(item.CreatedByName))
w.Write(topic_alt_55)
if item.Tag != "" {
w.Write(topic_alt_56)
w.Write([]byte(item.Tag))
w.Write(topic_alt_57)
} else {
w.Write(topic_alt_58)
w.Write([]byte(strconv.Itoa(item.Level)))
w.Write(topic_alt_59)
}
w.Write(topic_alt_60)
if item.ActionType != "" {
w.Write(topic_alt_61)
}
w.Write(topic_alt_62)
if item.ActionType != "" {
w.Write(topic_alt_63)
w.Write([]byte(item.ActionIcon))
w.Write(topic_alt_64)
w.Write([]byte(item.ActionType))
w.Write(topic_alt_65)
} else {
w.Write(topic_alt_66)
w.Write([]byte(item.ContentHtml))
w.Write(topic_alt_67)
if tmpl_topic_alt_vars.CurrentUser.Loggedin {
if tmpl_topic_alt_vars.CurrentUser.Perms.LikeItem {
w.Write(topic_alt_68)
w.Write([]byte(strconv.Itoa(item.ID)))
w.Write(topic_alt_69)
}
if tmpl_topic_alt_vars.CurrentUser.Perms.EditReply {
w.Write(topic_alt_70)
w.Write([]byte(strconv.Itoa(item.ID)))
w.Write(topic_alt_71)
}
if tmpl_topic_alt_vars.CurrentUser.Perms.DeleteReply {
w.Write(topic_alt_72)
w.Write([]byte(strconv.Itoa(item.ID)))
w.Write(topic_alt_73)
}
w.Write(topic_alt_74)
w.Write([]byte(strconv.Itoa(item.ID)))
w.Write(topic_alt_75)
w.Write([]byte(tmpl_topic_alt_vars.CurrentUser.Session))
w.Write(topic_alt_76)
if tmpl_topic_alt_vars.CurrentUser.Perms.ViewIPs {
w.Write(topic_alt_77)
w.Write([]byte(item.IPAddress))
w.Write(topic_alt_78)
}
}
w.Write(topic_alt_79)
w.Write([]byte(item.CreatedAt))
w.Write(topic_alt_80)
if item.LikeCount > 0 {
w.Write(topic_alt_81)
w.Write([]byte(strconv.Itoa(item.LikeCount)))
w.Write(topic_alt_82)
}
w.Write(topic_alt_83)
}
w.Write(topic_alt_84)
}
}
w.Write(topic_alt_85)
if tmpl_topic_alt_vars.CurrentUser.Perms.CreateReply {
w.Write(topic_alt_86)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Topic.ID)))
w.Write(topic_alt_87)
}
w.Write(topic_alt_88)
w.Write(footer_0)
if len(tmpl_topic_alt_vars.Header.Themes) != 0 {
for _, item := range tmpl_topic_alt_vars.Header.Themes {
if !item.HideFromThemes {
w.Write(footer_1)
w.Write([]byte(item.Name))
w.Write(footer_2)
if tmpl_topic_alt_vars.Header.ThemeName == item.Name {
w.Write(footer_3)
}
w.Write(footer_4)
w.Write([]byte(item.FriendlyName))
w.Write(footer_5)
}
}
}
w.Write(footer_6)
if tmpl_topic_alt_vars.Header.Widgets.RightSidebar != "" {
w.Write(footer_7)
w.Write([]byte(string(tmpl_topic_alt_vars.Header.Widgets.RightSidebar)))
w.Write(footer_8)
}
w.Write(footer_9)
}
