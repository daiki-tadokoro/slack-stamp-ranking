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
		Limit: 10,
		Types: []string{"public_channel", "private_channel"},
	})
	if err != nil {
		log.Fatalf("Error getting conversations: %v", err)
	}

	emojiUsage := make(map[string]int)

	// 各チャンネルのメッセージを取得しスタンプの使用状況を集計
	for _, channel := range channels {
		historyParams := slack.GetConversationHistoryParameters{
			ChannelID: channel.ID,
			Limit:     10,
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
