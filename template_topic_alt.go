// Code generated by Gosora. More below:
/* This file was automatically generated by the software. Please don't edit it as your changes may be overwritten at any moment. */
// +build !no_templategen
package main
import "io"
import "strconv"

func init() {
	template_topic_alt_handle = template_topic_alt
	//o_template_topic_alt_handle = template_topic_alt
	ctemplates = append(ctemplates,"topic_alt")
	tmpl_ptr_map["topic_alt"] = &template_topic_alt_handle
	tmpl_ptr_map["o_topic_alt"] = template_topic_alt
}

func template_topic_alt(tmpl_topic_alt_vars TopicPage, w io.Writer) {
w.Write(header_0)
w.Write([]byte(tmpl_topic_alt_vars.Title))
w.Write(header_1)
if len(tmpl_topic_alt_vars.Header.Stylesheets) != 0 {
for _, item := range tmpl_topic_alt_vars.Header.Stylesheets {
w.Write(header_2)
w.Write([]byte(item))
w.Write(header_3)
}
}
w.Write(header_4)
if len(tmpl_topic_alt_vars.Header.Scripts) != 0 {
for _, item := range tmpl_topic_alt_vars.Header.Scripts {
w.Write(header_5)
w.Write([]byte(item))
w.Write(header_6)
}
}
w.Write(header_7)
w.Write([]byte(tmpl_topic_alt_vars.CurrentUser.Session))
w.Write(header_8)
w.Write(menu_0)
w.Write([]byte(tmpl_topic_alt_vars.Header.Site.Name))
w.Write(menu_1)
if tmpl_topic_alt_vars.CurrentUser.Loggedin {
w.Write(menu_2)
w.Write([]byte(tmpl_topic_alt_vars.CurrentUser.Slug))
w.Write(menu_3)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.CurrentUser.ID)))
w.Write(menu_4)
if tmpl_topic_alt_vars.CurrentUser.Is_Super_Mod {
w.Write(menu_5)
}
w.Write(menu_6)
w.Write([]byte(tmpl_topic_alt_vars.CurrentUser.Session))
w.Write(menu_7)
} else {
w.Write(menu_8)
}
w.Write(menu_9)
w.Write(header_9)
if tmpl_topic_alt_vars.Header.Widgets.RightSidebar != "" {
w.Write(header_10)
}
w.Write(header_11)
if len(tmpl_topic_alt_vars.Header.NoticeList) != 0 {
for _, item := range tmpl_topic_alt_vars.Header.NoticeList {
w.Write(header_12)
w.Write([]byte(item))
w.Write(header_13)
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
if tmpl_topic_alt_vars.Topic.Is_Closed {
w.Write(topic_alt_11)
}
}
w.Write(topic_alt_12)
w.Write([]byte(tmpl_topic_alt_vars.Topic.Title))
w.Write(topic_alt_13)
if tmpl_topic_alt_vars.Topic.Is_Closed {
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
w.Write([]byte(tmpl_topic_alt_vars.Topic.UserSlug))
w.Write(topic_alt_21)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Topic.CreatedBy)))
w.Write(topic_alt_22)
w.Write([]byte(tmpl_topic_alt_vars.Topic.CreatedByName))
w.Write(topic_alt_23)
if tmpl_topic_alt_vars.Topic.Tag != "" {
w.Write(topic_alt_24)
w.Write([]byte(tmpl_topic_alt_vars.Topic.Tag))
w.Write(topic_alt_25)
} else {
w.Write(topic_alt_26)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Topic.Level)))
w.Write(topic_alt_27)
}
w.Write(topic_alt_28)
w.Write([]byte(tmpl_topic_alt_vars.Topic.Content))
w.Write(topic_alt_29)
w.Write([]byte(tmpl_topic_alt_vars.Topic.Content))
w.Write(topic_alt_30)
if tmpl_topic_alt_vars.CurrentUser.Loggedin {
if tmpl_topic_alt_vars.CurrentUser.Perms.LikeItem {
w.Write(topic_alt_31)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Topic.ID)))
w.Write(topic_alt_32)
}
if tmpl_topic_alt_vars.CurrentUser.Perms.EditTopic {
w.Write(topic_alt_33)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Topic.ID)))
w.Write(topic_alt_34)
}
if tmpl_topic_alt_vars.CurrentUser.Perms.DeleteTopic {
w.Write(topic_alt_35)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Topic.ID)))
w.Write(topic_alt_36)
}
if tmpl_topic_alt_vars.CurrentUser.Perms.PinTopic {
if tmpl_topic_alt_vars.Topic.Sticky {
w.Write(topic_alt_37)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Topic.ID)))
w.Write(topic_alt_38)
} else {
w.Write(topic_alt_39)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Topic.ID)))
w.Write(topic_alt_40)
}
}
w.Write(topic_alt_41)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Topic.ID)))
w.Write(topic_alt_42)
w.Write([]byte(tmpl_topic_alt_vars.CurrentUser.Session))
w.Write(topic_alt_43)
if tmpl_topic_alt_vars.CurrentUser.Perms.ViewIPs {
w.Write(topic_alt_44)
w.Write([]byte(tmpl_topic_alt_vars.Topic.IpAddress))
w.Write(topic_alt_45)
}
}
w.Write(topic_alt_46)
w.Write([]byte(tmpl_topic_alt_vars.Topic.CreatedAt))
w.Write(topic_alt_47)
if tmpl_topic_alt_vars.Topic.LikeCount > 0 {
w.Write(topic_alt_48)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Topic.LikeCount)))
w.Write(topic_alt_49)
}
w.Write(topic_alt_50)
if len(tmpl_topic_alt_vars.ItemList) != 0 {
for _, item := range tmpl_topic_alt_vars.ItemList {
w.Write(topic_alt_51)
if item.ActionType != "" {
w.Write(topic_alt_52)
}
w.Write(topic_alt_53)
w.Write([]byte(item.Avatar))
w.Write(topic_alt_54)
w.Write([]byte(item.UserSlug))
w.Write(topic_alt_55)
w.Write([]byte(strconv.Itoa(item.CreatedBy)))
w.Write(topic_alt_56)
w.Write([]byte(item.CreatedByName))
w.Write(topic_alt_57)
if item.Tag != "" {
w.Write(topic_alt_58)
w.Write([]byte(item.Tag))
w.Write(topic_alt_59)
} else {
w.Write(topic_alt_60)
w.Write([]byte(strconv.Itoa(item.Level)))
w.Write(topic_alt_61)
}
w.Write(topic_alt_62)
if item.ActionType != "" {
w.Write(topic_alt_63)
}
w.Write(topic_alt_64)
if item.ActionType != "" {
w.Write(topic_alt_65)
w.Write([]byte(item.ActionIcon))
w.Write(topic_alt_66)
w.Write([]byte(item.ActionType))
w.Write(topic_alt_67)
} else {
w.Write(topic_alt_68)
w.Write([]byte(item.ContentHtml))
w.Write(topic_alt_69)
if tmpl_topic_alt_vars.CurrentUser.Loggedin {
if tmpl_topic_alt_vars.CurrentUser.Perms.LikeItem {
w.Write(topic_alt_70)
w.Write([]byte(strconv.Itoa(item.ID)))
w.Write(topic_alt_71)
}
if tmpl_topic_alt_vars.CurrentUser.Perms.EditReply {
w.Write(topic_alt_72)
w.Write([]byte(strconv.Itoa(item.ID)))
w.Write(topic_alt_73)
}
if tmpl_topic_alt_vars.CurrentUser.Perms.DeleteReply {
w.Write(topic_alt_74)
w.Write([]byte(strconv.Itoa(item.ID)))
w.Write(topic_alt_75)
}
w.Write(topic_alt_76)
w.Write([]byte(strconv.Itoa(item.ID)))
w.Write(topic_alt_77)
w.Write([]byte(tmpl_topic_alt_vars.CurrentUser.Session))
w.Write(topic_alt_78)
if tmpl_topic_alt_vars.CurrentUser.Perms.ViewIPs {
w.Write(topic_alt_79)
w.Write([]byte(item.IpAddress))
w.Write(topic_alt_80)
}
}
w.Write(topic_alt_81)
w.Write([]byte(item.CreatedAt))
w.Write(topic_alt_82)
if item.LikeCount > 0 {
w.Write(topic_alt_83)
w.Write([]byte(strconv.Itoa(item.LikeCount)))
w.Write(topic_alt_84)
}
w.Write(topic_alt_85)
}
w.Write(topic_alt_86)
}
}
w.Write(topic_alt_87)
if tmpl_topic_alt_vars.CurrentUser.Perms.CreateReply {
w.Write(topic_alt_88)
w.Write([]byte(strconv.Itoa(tmpl_topic_alt_vars.Topic.ID)))
w.Write(topic_alt_89)
}
w.Write(footer_0)
if tmpl_topic_alt_vars.Header.Widgets.RightSidebar != "" {
w.Write(footer_1)
w.Write([]byte(string(tmpl_topic_alt_vars.Header.Widgets.RightSidebar)))
w.Write(footer_2)
}
w.Write(footer_3)
}
