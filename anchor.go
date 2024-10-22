package disc

import (
	"github.com/bwmarrin/discordgo"
)

type AnchorOpts struct {
	ForceClear         bool
	MaxAllowedMessages int
}

type AnchorOptFunc func(*AnchorOpts)

func DefaultAnchorOpts() *AnchorOpts {
	return &AnchorOpts{
		ForceClear:         false,
		MaxAllowedMessages: 1,
	}
}

func ForceClearAnchorOpt() AnchorOptFunc {
	return func(opts *AnchorOpts) {
		opts.ForceClear = true
	}
}

func MaxAllowedMessagesAnchorOpt(mam int) AnchorOptFunc {
	return func(opts *AnchorOpts) {
		opts.MaxAllowedMessages = mam
	}
}

func (c *Client) Anchor(channelID string, msg *discordgo.MessageSend, opts ...AnchorOptFunc) (err error) {
	o := DefaultAnchorOpts()
	for _, opt := range opts {
		opt(o)
	}
	validMsgs := 0
	for {
		msgs, err := c.sess.ChannelMessages(channelID, 100, "", "", "")
		if err != nil {
			return err
		}
		validMsgs = 0
		invaldMsgIDs := []string{}
		for _, m := range msgs {
			if o.ForceClear || m.Author.ID != c.sess.State.User.ID {
				invaldMsgIDs = append(invaldMsgIDs, m.ID)
			} else {
				validMsgs++
			}
		}
		if len(invaldMsgIDs) == 0 {
			break
		}
		err = c.anchorDeleteMessages(channelID, invaldMsgIDs)
		if err != nil {
			return err
		}
	}
	if validMsgs < o.MaxAllowedMessages {
		_, err = c.sess.ChannelMessageSendComplex(channelID, msg)
	}
	return
}

func (c *Client) anchorDeleteMessages(channelID string, invalidMsgIDs []string) (err error) {
	err = c.sess.ChannelMessagesBulkDelete(channelID, invalidMsgIDs)
	if err != nil {
		for _, msgID := range invalidMsgIDs {
			err = c.sess.ChannelMessageDelete(channelID, msgID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Client) Anchors(channelID string, msgs []*discordgo.MessageSend, opts ...AnchorOptFunc) (err error) {
	o := DefaultAnchorOpts()
	o.MaxAllowedMessages = len(msgs)
	for _, opt := range opts {
		opt(o)
	}
	validMsgs := 0
	for {
		msgs, err := c.sess.ChannelMessages(channelID, 100, "", "", "")
		if err != nil {
			return err
		}
		validMsgs = 0
		invaldMsgIDs := []string{}
		for _, m := range msgs {
			if o.ForceClear || m.Author.ID != c.sess.State.User.ID {
				invaldMsgIDs = append(invaldMsgIDs, m.ID)
			} else {
				validMsgs++
			}
		}
		if len(invaldMsgIDs) == 0 {
			break
		}
		err = c.sess.ChannelMessagesBulkDelete(channelID, invaldMsgIDs)
		if err != nil {
			return err
		}
	}
	for _, msg := range msgs {
		if validMsgs >= o.MaxAllowedMessages {
			break
		}
		_, err = c.sess.ChannelMessageSendComplex(channelID, msg)
		if err != nil {
			return err
		}
		validMsgs++
	}
	return
}
