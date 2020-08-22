package botutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	dg "github.com/bwmarrin/discordgo"
	"net/http"
)

// Embed structure
type Embed struct {
	*dg.MessageEmbed
}

// embedError error structure
type embedError struct {
	failingArgument string
	argumentLength int
	maxValue int
}

// Constants for message embed character limits
const (
	embedLimitTitle       = 256
	embedLimitDescription = 2048
	embedLimitFieldValue  = 1024
	embedLimitFieldName   = 256
	embedLimitField       = 25
	embedLimitFooter      = 2048
)

func (err *embedError) Error() string {
	return fmt.Sprintf("%s with length %v exceeded character limit of %v.",
		err.failingArgument, err.argumentLength, err.maxValue)
}

func VerifyEmbed(e *dg.MessageEmbed) error {
	if len(e.Title) > embedLimitTitle {
		return &embedError{"Title", len(e.Title), embedLimitTitle}
	}

	if len(e.Description) > embedLimitDescription {
		return &embedError{"Description", len(e.Description), embedLimitDescription}
	}

	if len(e.Fields) > embedLimitField {
		return &embedError{"Fields", len(e.Fields), embedLimitField}
	}

	if e.Footer != nil {
		if len(e.Footer.Text) > embedLimitFooter {
			return &embedError{"Footer", len(e.Footer.Text), embedLimitFooter}
		}
	}

	for i := range e.Fields {
		embedField := e.Fields[i]
		if len(embedField.Value) > embedLimitFieldValue {
			return &embedError{"Value", len(embedField.Value), embedLimitFieldValue}
		}

		if len(embedField.Name) > embedLimitFieldName {
			return &embedError{"Name", len(embedField.Name), embedLimitFieldName}
		}
	}

	return nil
}

// NewEmbed returns a new embed object
func NewEmbed() *Embed {
	return &Embed{&dg.MessageEmbed{}}
}

// SetTitle takes <Title string>
// Sets embed with title set <Title>
func (e *Embed) SetTitle(Title string) {
	e.Title = Title
}

// SetDescription takes <Description string>
// Sets embed with description set <Description>
func (e *Embed) SetDescription(Description string) {
	e.Description = Description
}

// AddField takes <Name string> <Value string> <Inline bool>
// Returns an embed with embed field set <Name> <Value> <Inline>
func (e *Embed) AddField(Name string, Value string, Inline bool) {
	EmbedField := dg.MessageEmbedField{
		Name:   Name,
		Value:  Value,
		Inline: Inline,
	}

	e.Fields = append(e.Fields, &EmbedField)
}

// SetFooter takes <IconURL string> <Text string>
// Sets embed with embed footer set <Iconurl> <Text>
func (e *Embed) SetFooter(IconURL string, Text string) {
	e.Footer = &dg.MessageEmbedFooter{
		IconURL: IconURL,
		Text:    Text,
	}
}

// SetImage takes <URL string>
// Sets embed image <URL>
func (e *Embed) SetImage(URL string) {
	e.Image = &dg.MessageEmbedImage{
		URL: URL,
	}
}

// SetThumbnail takes <URL string>
// Sets embed thumbnail to <URL>
func (e *Embed) SetThumbnail(URL string) {
	e.Thumbnail = &dg.MessageEmbedThumbnail{
		URL: URL,
	}
}

// SetAuthor takes <Name string> <IconURL string> <URL string>
// Returns embed with author set <Name> <IconURL> <URL>
func (e *Embed) SetAuthor(Name string, IconURL string, URL string) {
	e.Author = &dg.MessageEmbedAuthor{
		Name:    Name,
		IconURL: IconURL,
		URL:     URL,
	}
}

// SetURL takes <URL string>
// Sets embed url <URL>
func (e *Embed) SetURL(URL string) {
	e.URL = URL
}

// SetColor takes <Color string>
// Sets color of embed to <Color>
// Returns error
func (e *Embed) SetColor(ColorByte int) {
	e.Color = ColorByte
}

// InlineAllFields sets all fields in the embed to be inline
func (e *Embed) InlineAllFields() *Embed {
	for _, v := range e.Fields {
		v.Inline = true
	}
	return e
}

// Truncate truncates the number of embed fields over the character limit.
func (e *Embed) Truncate() *Embed {
	e.truncateFields()
	return e
}

// TruncateFields truncates fields that are too long
func (e *Embed) truncateFields() *Embed {
	if len(e.Fields) > embedLimitField {
		e.Fields = e.Fields[:embedLimitField]
	}
	return e
}

// SendToWebhook takes <Webhook string>
// Sends embed to webhook
// Returns error if invalid embed or error posting to webhook
func (e *Embed) SendToWebhook(Webhook string) error {
	err := VerifyEmbed(e.MessageEmbed)

	if err != nil {
		return err
	}

	embedArray := append(make([]*dg.MessageEmbed, 0), e.MessageEmbed)
	params := dg.WebhookParams{
		Embeds: embedArray,
	}

	embedJSON, err := json.Marshal(params)
	if err != nil {
		return err
	}

	_, err = http.Post(Webhook, "application/json", bytes.NewBuffer(embedJSON))
	if err != nil {
		return err
	}

	return nil
}

// SendToChannel takes <s *dg.Session>, <channelID string>
// Verifies embed and sends embed to channel
// Returns message and error
func (e *Embed) SendToChannel(s *dg.Session, channelID string) (*dg.Message, error) {
	err := VerifyEmbed(e.MessageEmbed)

	if err != nil {
		return nil, err
	}

	return s.ChannelMessageSendComplex(channelID, &dg.MessageSend{
		Embed: e.MessageEmbed,
	})
}

// ChannelMessageEditEmbed takes <s *dg.Session>, <channelID string>, <messageID string>
// Verifies embed and edits message with new embed
// Returns message and error
func (e *Embed) ChannelMessageEditEmbed(s *dg.Session, channelID string, messageID string) (*dg.Message, error) {
	err := VerifyEmbed(e.MessageEmbed)

	if err != nil {
		return nil, err
	}

	return s.ChannelMessageEditComplex(&dg.MessageEdit{
		Embed:           e.MessageEmbed,
		ID:              messageID,
		Channel:         channelID,
	})
}
