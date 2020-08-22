package botutil

import (
	"fmt"
	dg "github.com/bwmarrin/discordgo"
	"reflect"
	"time"
)

// ControlEmojis contains the structure and default structure for pagination emojis
type ControlEmojis struct {
	toBegin string `default:"⏮"`
	backwards string `default:"⏪"`
	forwards string `default:"⏩"`
	toEnd string `default:"⏭"`
	stop string `default:"⏹"`
}

// Paginator contains the structure for the paginator.
type Paginator struct {
	// The sent message that hosts the paginator
	message *dg.Message
	// The channel the paginator was sent in
	channelID string
	// The pages of the paginator
	pages []*dg.MessageEmbed
	// The emojis used to control the paginator
	controlEmojis ControlEmojis
	// Current page the paginator is on
	index int
	// The discord session it uses EoA
	s *dg.Session
	// If reactions are deleted
	reactionsRemoved bool
	// User the paginator is locked to
	user *dg.User
	// Time after which the paginator is disabled
	timeOut time.Duration
	// If paginator is still active
	active bool
	// Print information messages
	infoMessages bool
}

// paginatorError error structure
type paginatorError struct {
	failingFunction string
	failingReason string
}

func (err *paginatorError) Error() string {
	return fmt.Sprintf("Function %s failed because:\n %s", err.failingFunction, err.failingReason)
}

// NewPaginator takes <s *dg.Session>, <channelID string>, <user *dg.User>
// Returns standard *Paginator
func NewPaginator(s *dg.Session, channelID string, user *dg.User, controlEmojis ControlEmojis,
	timeOut time.Duration, infoMessages bool) *Paginator {
	p := &Paginator{
		message:          nil,
		channelID:        channelID,
		pages:            nil,
		controlEmojis:    controlEmojis,
		index:            0,
		s:                s,
		reactionsRemoved: false,
		user:             user,
		timeOut:          timeOut,
		active: 		  false,
		infoMessages: 	  infoMessages,
	}

	return p
}

// Add takes <e *dg.MessageEmbed>
// Verifies embed and adds embed to the paginator pages
// returns error
func (p *Paginator) Add(e *dg.MessageEmbed) error {
	err := VerifyEmbed(e)

	if err != nil {
		return err
	}

	p.pages = append(p.pages, e)

	return nil
}

// Run runs the paginator
func (p *Paginator) Run() error {
	if p.active {
		return &paginatorError{
			failingFunction: "Run",
			failingReason:   "Paginator is already running.",
		}
	}

	if len(p.pages) == 0 {
		return &paginatorError{
			failingFunction: "Run",
			failingReason:   "No pages found in paginator.pages",
		}
	}

	err := p.setPageNumber()

	if err != nil {
		return err
	}

	msg, err := p.s.ChannelMessageSendComplex(p.channelID, &dg.MessageSend{
		Embed: p.pages[p.index],
	})

	if err != nil {
		return &paginatorError{
			failingFunction: "Run",
			failingReason:   err.Error(),
		}
	}

	start := time.Now()
	p.message = msg
	p.active = true

	err = p.addReactions()

	if err != nil {
		return err
	}

	var reaction *dg.MessageReaction

	for {
		select {
		case reactionEvent := <-p.nextReaction():
			reaction = reactionEvent.MessageReaction
		case <-time.After(start.Add(p.timeOut).Sub(time.Now())):
			return p.close()
		}

		if reaction.MessageID != p.message.ID || !p.isPaginatorUser(reaction.UserID) || reaction.UserID == "" {
			continue
		}

		go func() {
			switch reaction.Emoji.Name {
			case p.controlEmojis.toBegin:
				err = p.firstPage()
			case p.controlEmojis.backwards:
				err = p.previousPage()
			case p.controlEmojis.forwards:
				err = p.nextPage()
			case p.controlEmojis.toEnd:
				err = p.lastPage()
			case p.controlEmojis.stop:
				err = p.close()
			}
		}()

		if err != nil {
			return err
		}

		go func() {
			// is allowed to error freely. This is to prevent people spamming reactions.
			time.Sleep(time.Millisecond*100)
			_ = p.s.MessageReactionRemove(reaction.ChannelID, reaction.MessageID, reaction.Emoji.Name, reaction.UserID)
		}()
	}
}

// nextReaction activates message reaction handler
// awaits reaction and returns it
func (p *Paginator) nextReaction() chan *dg.MessageReactionAdd {
	channel := make(chan *dg.MessageReactionAdd)
	p.s.AddHandlerOnce(func(_ *dg.Session, reaction *dg.MessageReactionAdd) {
		channel <- reaction
	})
	return channel
}

// addReactions is an internal function that adds the pagination reactions to the paginator.
// returns error
func (p *Paginator) addReactions() error {
	p.fillDefaultReactions()

	for i := 0; i< reflect.ValueOf(p.controlEmojis).NumField(); i++ {
		err := p.s.MessageReactionAdd(p.channelID, p.message.ID, reflect.ValueOf(p.controlEmojis).Field(i).String())

		if err != nil {
			return err
		}
	}

	return nil
}

// nextPage is an internal function that increases the index
// returns error
func (p *Paginator) nextPage() error {
	if p.index+1 == len(p.pages) {
		if p.infoMessages {
			fmt.Println(&paginatorError{
				failingFunction: "nextPage",
				failingReason:   "Page index is already max.",
			})
		}
		return nil
	}

	p.index += 1

	err := p.updatePaginatorMessage(p.pages[p.index])

	return err
}

// previousPage is an internal function that decreases the index
// returns error
func (p *Paginator) previousPage() error {
	if p.index == 0 {
		if p.infoMessages {
			fmt.Println(&paginatorError{
				failingFunction: "previousPage",
				failingReason:   "Page index is north.",
			})
		}
		return nil
	}

	p.index -= 1

	return p.updatePaginatorMessage(p.pages[p.index])
}

// firstPage is an internal function that resets to the first page
// returns error
func (p *Paginator) firstPage() error {
	if p.index == 0 {
		if p.infoMessages {
			fmt.Println(&paginatorError{
				failingFunction: "previousPage",
				failingReason:   "Page index is north.",
			})
		}
		return nil
	}

	p.index = 0

	return p.updatePaginatorMessage(p.pages[p.index])
}

// lastPage is an internal function that resets to the first page
// returns error
func (p *Paginator) lastPage() error {
	if p.index+1 == len(p.pages) {
		if p.infoMessages {
			fmt.Println(&paginatorError{
				failingFunction: "lastPage",
				failingReason:   "Page index is already max.",
			})
		}
		return nil
	}

	p.index = len(p.pages)-1

	return p.updatePaginatorMessage(p.pages[p.index])
}

// updateMessage takes <e *dg.MessageEmbed>
// updates the message in the paginator that is currently displayed
// returns error
func (p *Paginator) updatePaginatorMessage(e *dg.MessageEmbed) error {
	_, err := p.s.ChannelMessageEditComplex(&dg.MessageEdit{
		Embed:           e,
		ID:              p.message.ID,
		Channel:         p.channelID,
	})

	return err
}

// close puts the paginator as inactive
// removes all reactions to the paginator
func (p *Paginator) close() error {
	p.active = false
	if p.reactionsRemoved {
		return nil
	}

	err := p.s.MessageReactionsRemoveAll(p.channelID, p.message.ID)

	if err == nil {
		p.reactionsRemoved = true
	}

	return err
}

// isPaginatorUser takes <userID string>
// checks if user is allowed to interact with paginator
// returns boolean
func (p *Paginator) isPaginatorUser(userID string) bool {
	if userID == p.user.ID {
		return true
	}

	return false
}

// setPageNumber sets page number on all pages
// returns error
func (p *Paginator) setPageNumber() error {
	if len(p.pages) == 0 {
		return &paginatorError{
			failingFunction: "setPageNumber",
			failingReason:   "No pages found in paginator.pages",
		}
	}

	for i, s := range p.pages {
		s.Footer = &dg.MessageEmbedFooter{
			Text:         fmt.Sprintf("%d/%d", i+1, len(p.pages)),
		}
	}

	return nil
}

// fillDefaultReactions checks if controlEmoji is empty
// sets default emojis if controlEmoji is empty
func (p *Paginator) fillDefaultReactions() {
	if p.controlEmojis.toBegin == "" { p.controlEmojis.toBegin = "⏮" }
	if p.controlEmojis.backwards == "" { p.controlEmojis.backwards = "⏪" }
	if p.controlEmojis.forwards == "" { p.controlEmojis.forwards = "⏩" }
	if p.controlEmojis.toEnd == "" { p.controlEmojis.toEnd = "⏭" }
	if p.controlEmojis.stop == "" { p.controlEmojis.stop = "⏹" }
}
