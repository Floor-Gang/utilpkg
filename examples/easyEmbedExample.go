package utilpkg

import (
	"github.com/Floor-Gang/utilpkg/botutil"
	"log"
)

// main shows an example of making an embed
//  Using the embed utils in utils/easyEmbed.go
func main() {
	//Create a new embed object
	embed := botutil.NewEmbed()
	embed.SetTitle("Title")
	embed.SetDescription("Description")
	embed.SetURL("URL")
	embed.AddField("Name", "Value", true)
	embed.SetThumbnail("URL")
	embed.SetImage("URL")
	err := embed.SetColor("#F1B379")

	if err != nil {
		log.Println(err)
	}

	embed.SetAuthor("Name", "URL", "IconURL")
	embed.SetFooter("Text", "IconURL")
	embed.Truncate()
	err = embed.SendToWebhook("Webhook URL")

	if err != nil {
		log.Println(err)
	}
}
