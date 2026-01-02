//go:build ignore

package main

import (
	"chat_backend/internal/model"

	"gorm.io/gen"
)

func main() {
	g := gen.NewGenerator(gen.Config{
		OutPath: "../dao",
		Mode:    gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface,
	})

	//g.UseDB(nil) // 不连接真实 DB，仅基于结构体生成

	g.ApplyBasic(
		model.User{},
		model.Friend{},
		model.Group{},
		model.GroupMember{},
		model.Message{},
		model.InvitationCode{},
		model.FriendRequest{},
		model.GroupJoinRequest{},
		model.MessageReceipt{},
	)

	g.Execute()
}
