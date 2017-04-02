// Code generated by. DO NOT EDIT.
/* This file was automatically generated by the software. Please don't edit it as your changes may be overwritten at any moment. */
package main
import "io"
import "strconv"

func init() {
	template_topic_handle = template_topic
	//o_template_topic_handle = template_topic
	ctemplates = append(ctemplates,"topic")
	tmpl_ptr_map["topic"] = &template_topic_handle
	tmpl_ptr_map["o_topic"] = template_topic
}

func template_topic(tmpl_topic_vars TopicPage, w io.Writer) {
w.Write(header_0)
w.Write([]byte(tmpl_topic_vars.Title))
w.Write(header_1)
w.Write([]byte(tmpl_topic_vars.CurrentUser.Session))
w.Write(header_2)
w.Write(menu_0)
if tmpl_topic_vars.CurrentUser.Loggedin {
w.Write(menu_1)
w.Write([]byte(strconv.Itoa(tmpl_topic_vars.CurrentUser.ID)))
w.Write(menu_2)
if tmpl_topic_vars.CurrentUser.Is_Super_Mod {
w.Write(menu_3)
}
w.Write(menu_4)
w.Write([]byte(tmpl_topic_vars.CurrentUser.Session))
w.Write(menu_5)
} else {
w.Write(menu_6)
}
w.Write(menu_7)
w.Write(header_3)
if len(tmpl_topic_vars.NoticeList) != 0 {
for _, item := range tmpl_topic_vars.NoticeList {
w.Write(header_4)
w.Write([]byte(item))
w.Write(header_5)
}
}
if tmpl_topic_vars.Page > 1 {
w.Write(topic_0)
w.Write([]byte(strconv.Itoa(tmpl_topic_vars.Topic.ID)))
w.Write(topic_1)
w.Write([]byte(strconv.Itoa(tmpl_topic_vars.Page - 1)))
w.Write(topic_2)
}
if tmpl_topic_vars.LastPage != tmpl_topic_vars.Page {
w.Write(topic_3)
w.Write([]byte(strconv.Itoa(tmpl_topic_vars.Topic.ID)))
w.Write(topic_4)
w.Write([]byte(strconv.Itoa(tmpl_topic_vars.Page + 1)))
w.Write(topic_5)
w.Write([]byte(strconv.Itoa(tmpl_topic_vars.Topic.ID)))
w.Write(topic_6)
w.Write([]byte(strconv.Itoa(tmpl_topic_vars.Page + 1)))
w.Write(topic_7)
}
w.Write(topic_8)
w.Write([]byte(strconv.Itoa(tmpl_topic_vars.Topic.ID)))
w.Write(topic_9)
if tmpl_topic_vars.Topic.Sticky {
w.Write(topic_10)
} else {
if tmpl_topic_vars.Topic.Is_Closed {
w.Write(topic_11)
}
}
w.Write(topic_12)
w.Write([]byte(tmpl_topic_vars.Topic.Title))
w.Write(topic_13)
if tmpl_topic_vars.Topic.Is_Closed {
w.Write(topic_14)
}
if tmpl_topic_vars.CurrentUser.Perms.EditTopic {
w.Write(topic_15)
w.Write([]byte(tmpl_topic_vars.Topic.Title))
w.Write(topic_16)
if tmpl_topic_vars.CurrentUser.Perms.CloseTopic {
w.Write(topic_17)
}
w.Write(topic_18)
}
w.Write(topic_19)
if tmpl_topic_vars.Topic.Avatar != "" {
w.Write(topic_20)
w.Write([]byte(tmpl_topic_vars.Topic.Avatar))
w.Write(topic_21)
if tmpl_topic_vars.Topic.ContentLines <= 5 {
w.Write(topic_22)
}
w.Write(topic_23)
w.Write([]byte(string(tmpl_topic_vars.Topic.Css)))
}
w.Write(topic_24)
w.Write([]byte(tmpl_topic_vars.Topic.Content))
w.Write(topic_25)
w.Write([]byte(tmpl_topic_vars.Topic.Content))
w.Write(topic_26)
w.Write([]byte(strconv.Itoa(tmpl_topic_vars.Topic.CreatedBy)))
w.Write(topic_27)
w.Write([]byte(tmpl_topic_vars.Topic.CreatedByName))
w.Write(topic_28)
if tmpl_topic_vars.CurrentUser.Perms.LikeItem {
w.Write(topic_29)
w.Write([]byte(strconv.Itoa(tmpl_topic_vars.Topic.ID)))
w.Write(topic_30)
if tmpl_topic_vars.Topic.Liked {
w.Write(topic_31)
}
w.Write(topic_32)
}
if tmpl_topic_vars.CurrentUser.Perms.EditTopic {
w.Write(topic_33)
w.Write([]byte(strconv.Itoa(tmpl_topic_vars.Topic.ID)))
w.Write(topic_34)
}
if tmpl_topic_vars.CurrentUser.Perms.DeleteTopic {
w.Write(topic_35)
w.Write([]byte(strconv.Itoa(tmpl_topic_vars.Topic.ID)))
w.Write(topic_36)
}
if tmpl_topic_vars.CurrentUser.Perms.PinTopic {
if tmpl_topic_vars.Topic.Sticky {
w.Write(topic_37)
w.Write([]byte(strconv.Itoa(tmpl_topic_vars.Topic.ID)))
w.Write(topic_38)
} else {
w.Write(topic_39)
w.Write([]byte(strconv.Itoa(tmpl_topic_vars.Topic.ID)))
w.Write(topic_40)
}
}
w.Write(topic_41)
w.Write([]byte(strconv.Itoa(tmpl_topic_vars.Topic.ID)))
w.Write(topic_42)
w.Write([]byte(tmpl_topic_vars.CurrentUser.Session))
w.Write(topic_43)
if tmpl_topic_vars.Topic.LikeCount > 0 {
w.Write(topic_44)
w.Write([]byte(strconv.Itoa(tmpl_topic_vars.Topic.LikeCount)))
w.Write(topic_45)
}
if tmpl_topic_vars.Topic.Tag != "" {
w.Write(topic_46)
w.Write([]byte(tmpl_topic_vars.Topic.Tag))
w.Write(topic_47)
} else {
w.Write(topic_48)
w.Write([]byte(strconv.Itoa(tmpl_topic_vars.Topic.Level)))
w.Write(topic_49)
}
w.Write(topic_50)
if len(tmpl_topic_vars.ItemList) != 0 {
for _, item := range tmpl_topic_vars.ItemList {
if item.ActionType != "" {
w.Write(topic_51)
w.Write([]byte(item.ActionIcon))
w.Write(topic_52)
w.Write([]byte(item.ActionType))
w.Write(topic_53)
} else {
w.Write(topic_54)
if item.Avatar != "" {
w.Write(topic_55)
w.Write([]byte(item.Avatar))
w.Write(topic_56)
if item.ContentLines <= 5 {
w.Write(topic_57)
}
w.Write(topic_58)
w.Write([]byte(string(item.Css)))
}
w.Write(topic_59)
w.Write([]byte(item.ContentHtml))
w.Write(topic_60)
w.Write([]byte(strconv.Itoa(item.CreatedBy)))
w.Write(topic_61)
w.Write([]byte(item.CreatedByName))
w.Write(topic_62)
if tmpl_topic_vars.CurrentUser.Perms.LikeItem {
w.Write(topic_63)
w.Write([]byte(strconv.Itoa(item.ID)))
w.Write(topic_64)
if item.Liked {
w.Write(topic_65)
}
w.Write(topic_66)
}
if tmpl_topic_vars.CurrentUser.Perms.EditReply {
w.Write(topic_67)
w.Write([]byte(strconv.Itoa(item.ID)))
w.Write(topic_68)
}
if tmpl_topic_vars.CurrentUser.Perms.DeleteReply {
w.Write(topic_69)
w.Write([]byte(strconv.Itoa(item.ID)))
w.Write(topic_70)
}
w.Write(topic_71)
w.Write([]byte(strconv.Itoa(item.ID)))
w.Write(topic_72)
w.Write([]byte(tmpl_topic_vars.CurrentUser.Session))
w.Write(topic_73)
if item.LikeCount > 0 {
w.Write(topic_74)
w.Write([]byte(strconv.Itoa(item.LikeCount)))
w.Write(topic_75)
}
if item.Tag != "" {
w.Write(topic_76)
w.Write([]byte(item.Tag))
w.Write(topic_77)
} else {
w.Write(topic_78)
w.Write([]byte(strconv.Itoa(item.Level)))
w.Write(topic_79)
}
w.Write(topic_80)
}
}
}
w.Write(topic_81)
if tmpl_topic_vars.CurrentUser.Perms.CreateReply {
w.Write(topic_82)
w.Write([]byte(strconv.Itoa(tmpl_topic_vars.Topic.ID)))
w.Write(topic_83)
}
w.Write(footer_0)
}
