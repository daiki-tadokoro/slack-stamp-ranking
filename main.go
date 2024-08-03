package main

import (
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
)

func main() {
	// .envファイルを読み込む
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// 環境変数を取得
	slackToken := os.Getenv("SLACK_API_TOKEN")
	if slackToken == "" {
		log.Fatal("SLACK_API_TOKEN must be set")
	}

	api := slack.New(slackToken)

	// スコープを確認
	authTest, err := api.AuthTest()
	if err != nil {
		log.Fatalf("Error during AuthTest: %v", err)
	}
	fmt.Printf("Authenticated as user: %s\n", authTest.User)

	// 全てのチャンネルのリストを取得
	channels, _, err := api.GetConversations(&slack.GetConversationsParameters{
		Limit: 100,
		Types: []string{"public_channel", "private_channel"},
	})
	if err != nil {
		log.Fatalf("Error getting conversations: %v", err)
	}

	emojiUsage := make(map[string]int)

	// 各チャンネルにBotを追加
	for _, channel := range channels {
		if channel.IsArchived {
			log.Printf("Skipping archived channel: %s\n", channel.Name)
			continue
		}

		// チャンネルに参加する
		_, _, _, err := api.JoinConversation(channel.ID)
		if err != nil {
			log.Printf("Error joining channel %s: %v", channel.Name, err)
			// 既に参加しているか確認
			if !isAlreadyInChannel(api, channel.ID) {
				continue
			}
		}
		fmt.Printf("Joined channel: %s\n", channel.Name)

		// 各チャンネルのメッセージを取得しスタンプの使用状況を集計
		historyParams := slack.GetConversationHistoryParameters{
			ChannelID: channel.ID,
			Limit:     100,
		}

		history, err := api.GetConversationHistory(&historyParams)
		if err != nil {
			log.Printf("Error getting conversation history for channel %s: %v", channel.Name, err)
			continue
		}

		for _, message := range history.Messages {
			for _, reaction := range message.Reactions {
				emojiUsage[reaction.Name] += reaction.Count
			}
		}
	}

	// スタンプの使用数でソート
	type emojiCount struct {
		Name  string
		Count int
	}
	var emojiList []emojiCount
	for emoji, count := range emojiUsage {
		emojiList = append(emojiList, emojiCount{Name: emoji, Count: count})
	}
	sort.Slice(emojiList, func(i, j int) bool {
		return emojiList[i].Count > emojiList[j].Count
	})

	fmt.Println("Emoji Usage Rankings:")
	for i, emoji := range emojiList {
		fmt.Printf("%d. %s: %d\n", i+1, emoji.Name, emoji.Count)
	}
}

// チャンネルに既に参加しているかを確認するヘルパー関数
func isAlreadyInChannel(api *slack.Client, channelID string) bool {
	channelInfo, err := api.GetConversationInfo(&slack.GetConversationInfoInput{
		ChannelID: channelID,
	})
	if err != nil {
		log.Printf("Error getting conversation info for channel %s: %v", channelID, err)
		return false
	}
	return channelInfo.IsMember
}
