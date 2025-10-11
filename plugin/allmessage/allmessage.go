package broadcast

import (
    "fmt"
    "strings"
    "time"

    ctrl "github.com/FloatTech/zbpctrl"
    "github.com/FloatTech/zbputils/control"
    "github.com/FloatTech/zbputils/ctxext"
    "github.com/wdvxdr1123/ZeroBot"
    "github.com/wdvxdr1123/ZeroBot/message"
)

// 黑名单群列表
var blacklist = map[int64]bool{
    123456789: true,
    987654321: true,
}

// 限速器
var limiter = ctxext.NewLimiterManager(time.Second*10, 1)

func init() {
    engine := control.Register("broadcast", &ctrl.Options[*zero.Ctx]{
        DisableOnDefault: false,
        Brief:            "群广播插件",
        Help:             "- 广播(预览) 图片+文本 或纯文本\n- 支持图文混合广播\n- 支持预览模式",
        PublicDataFolder: "Broadcast",
        OnEnable: func(ctx *zero.Ctx) {
            ctx.Send("广播插件已启用")
        },
        OnDisable: func(ctx *zero.Ctx) {
            ctx.Send("广播插件已禁用")
        },
    })

    engine.OnRegex(`^广播(预览)?\s+(.+)$`).SetBlock(true).Limit(limiter.LimitByGroup).Handle(func(ctx *zero.Ctx) {
        matches := ctx.State["regex_matched"].([]string)
        isPreview := matches[1] == "预览"
        raw := matches[2]

        var msg message.Slice
        parts := strings.Fields(raw)

        if len(parts) >= 2 && parts[0] == "图片" {
            for i := 1; i < len(parts); i++ {
                part := parts[i]
                if strings.HasPrefix(part, "http") {
                    msg = append(msg, message.Image(part))
                } else {
                    text := strings.Join(parts[i:], " ")
                    msg = append(msg, message.Text("\n" + text))
                    break
                }
            }
        } else {
            msg = message.Slice{message.Text(raw)}
        }

        var nodes []message.Node
        var successCount, failCount, skipCount int

        for _, group := range ctx.Bot.GroupList() {
            if blacklist[group.GroupID] {
                skipCount++
                continue
            }

            status := "（预览中，不会真的发出去哟~）"
            if !isPreview {
                err := ctx.SendGroupMessage(group.GroupID, msg)
                if err != nil {
                    status = "呜呜，失败了……"
                    failCount++
                } else {
                    status = "发送成功啦！✨"
                    successCount++
                }
         }

            nodes = append(nodes, message.Node{
                SenderID:   ctx.SelfID,
                SenderName: ctx.Bot.Nickname,
                Content:    message.Text(fmt.Sprintf("群 %d：%s", group.GroupID, status)),
            })
        }

        if ctx.IsGroup {
            if isPreview {
                ctx.Send(message.Text("来看看广播预览吧～不是真的发出去哦！"))
            }
            ctx.SendGroupForwardMessage(ctx.GroupID, nodes)
        } else {
            if isPreview {
                ctx.Send(message.Text("这是广播预览结果哒～"))
                ctx.SendPrivateForwardMessage(ctx.UserID, nodes)
            } else {
                ctx.Send(message.Text(fmt.Sprintf(
                    "广播完成啦～\n成功发送 %d 个群\n失败 %d 个群\n跳过 %d 个黑名单群\n辛苦我啦！(๑•̀ㅂ•́)و✧",
                    successCount, failCount, skipCount,
                )))
            }
        }
    })
}
