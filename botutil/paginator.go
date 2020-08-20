package botutil

import (
	"fmt"
	dg "github.com/bwmarrin/discordgo"
	"reflect"
	"time"
)

// ControlEmojis contains the structure and default structure for pagination emojis
type ControlEmojis struct {
	toBegin string `default:":tack_previous:"`
	backwards string `default:":rewind:"`
	forwards string `default:":fast_forward:"`
	toEnd string `default:":track_next:"`
	stop string `default:":stop_button:"`
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
}

// paginatorError error structure
type paginatorError struct {
	failingFunction string
	failingReason string
}

func (err *paginatorError) Error() string {
	return fmt.Sprintf("Function %s failed because:\n %s", err.failingFunction, err.failingReason)
}

//// External commands
// NewPaginator | create new paginator						| ✓
// paginator.Add | add embed to paginator					| ✓
// paginator.Run | run paginator and activate handlers		|


// NewPaginator takes <s *dg.Session>, <channelID string>, <user *dg.User>
// Returns standard *Paginator
func NewPaginator(s *dg.Session, channelID string, user *dg.User, controlEmojis ControlEmojis,
	timeOut time.Duration) *Paginator {
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
	}

	return p
}

// Add takes <e *dg.MessageEmbed>
// Verifies embed and adds embed to the paginator pages
// returns error
func (p *Paginator) Add(e *dg.MessageEmbed) error {
	err := verifyEmbed(e)

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

	for {
		select {
		case <-time.After(start.Add(p.timeOut).Sub(time.Now())):
			return p.close()
		}
	}
}

//// Internal commands
// addReactions | add pagination reactions											| ✓
// nextPage | which reaction add and prepare message to update						| ✓
// previousPage | which reaction add and prepare message to update					| ✓
// firstPage | go to first page														| ✓
// lastPage | go to last page														| ✓
// updatePaginatorMessage | on reaction add											| ✓
// close | ControlEmojis.stop														| ✓
// isPaginatorUser | returns boolean if user is allowed to interact with paginator 	| ✓
// pageNumber | sets page number													| ✓

// addReactions is an internal function that adds the pagination reactions to the paginator.
// returns error
func (p *Paginator) addReactions() error {
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
	if p.index == len(p.pages)+1 {
		return &paginatorError{
			failingFunction: "nextPage",
			failingReason:   "Page index is already max.",
		}
	}

	p.index += 1

	return p.updatePaginatorMessage(p.pages[p.index])
}

// previousPage is an internal function that decreases the index
// returns error
func (p *Paginator) previousPage() error {
	if p.index == 0 {
		return &paginatorError{
			failingFunction: "previousPage",
			failingReason:   "Page index is north.",
		}
	}

	p.index -= 1

	return p.updatePaginatorMessage(p.pages[p.index])
}

// firstPage is an internal function that resets to the first page
// returns error
func (p *Paginator) firstPage() error {
	if p.index == 0 {
		return &paginatorError{
			failingFunction: "previousPage",
			failingReason:   "Page index is north.",
		}
	}

	p.index = 0

	return p.updatePaginatorMessage(p.pages[p.index])
}

// lastPage is an internal function that resets to the first page
// returns error
func (p *Paginator) lastPage() error {
	if p.index == 0 {
		return &paginatorError{
			failingFunction: "previousPage",
			failingReason:   "Page index is already max.",
		}
	}

	p.index = len(p.pages)-1

	return p.updatePaginatorMessage(p.pages[p.index])
}

// updateMessage takes <e *dg.MessageEmbed>
// updates the message in the paginator that is currently displayed
// returns error
func (p *Paginator) updatePaginatorMessage(e *dg.MessageEmbed) error {
	_, err := p.s.ChannelMessageEditComplex(dg.NewMessageEdit(p.message.ID, p.channelID).SetEmbed(e))

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

// isPaginatorUser takes <user dg.User>
// checks if user is allowed to interact with paginator
// returns boolean
func (p *Paginator) isPaginatorUser(user *dg.User) bool {
	if user == p.user {
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

	for _, s := range p.pages {
		s.Footer = &dg.MessageEmbedFooter{
			Text:         fmt.Sprintf("%d/%d", p.index+1, len(p.pages)),
		}
	}

	return nil
}
