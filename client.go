package disc

import (
	"github.com/bwmarrin/discordgo"
	"github.com/stevo-go-utils/structures"
)

type Client struct {
	sess                      *discordgo.Session
	appID                     string
	pingHandlers              *structures.SafeMap[string, PingHandler]
	appCmdHandlers            *structures.SafeMap[string, AppCmdHandler]
	msgComponentHandlers      *structures.SafeMap[string, MsgComponentHandler]
	appCmdAutoHandlers        *structures.SafeMap[string, AppCmdAutoHandler]
	modalSubmitHandlers       *structures.SafeMap[string, ModalSubmitHandler]
	pingHandlerErrChs         *structures.SafeMap[string, chan error]
	appCmdHandlerErrChs       *structures.SafeMap[string, chan error]
	msgComponentHandlerErrChs *structures.SafeMap[string, chan error]
	appCmdAutoHandlerErrChs   *structures.SafeMap[string, chan error]
	modalSubmitHandlerErrChs  *structures.SafeMap[string, chan error]
	handlerErrCh              chan error
}

func NewClient(token string, appID string) (c *Client, err error) {
	sess, err := discordgo.New("Bot " + token)
	if err != nil {
		return c, err
	}
	sess.Identify.Intents = discordgo.IntentsAll
	return &Client{
		sess:                      sess,
		appID:                     appID,
		pingHandlers:              structures.NewSafeMap[string, PingHandler](),
		appCmdHandlers:            structures.NewSafeMap[string, AppCmdHandler](),
		msgComponentHandlers:      structures.NewSafeMap[string, MsgComponentHandler](),
		appCmdAutoHandlers:        structures.NewSafeMap[string, AppCmdAutoHandler](),
		modalSubmitHandlers:       structures.NewSafeMap[string, ModalSubmitHandler](),
		pingHandlerErrChs:         structures.NewSafeMap[string, chan error](),
		appCmdHandlerErrChs:       structures.NewSafeMap[string, chan error](),
		msgComponentHandlerErrChs: structures.NewSafeMap[string, chan error](),
		appCmdAutoHandlerErrChs:   structures.NewSafeMap[string, chan error](),
		modalSubmitHandlerErrChs:  structures.NewSafeMap[string, chan error](),
		handlerErrCh:              nil,
	}, nil
}

func (c *Client) Handle() {
	c.sess.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionPing:
			if handler, ok := c.pingHandlers.Get(i.ApplicationCommandData().Name); ok {
				err := handler(BaseHandlerData{C: c, S: s, I: i})
				go func() {
					if err == nil {
						return
					}
					handlerErrCh, ok := c.pingHandlerErrChs.Get(i.ApplicationCommandData().Name)
					if ok {
						handlerErrCh <- err
					} else if c.handlerErrCh != nil {
						c.handlerErrCh <- err
					}
				}()
			}
		case discordgo.InteractionApplicationCommand:
			if handler, ok := c.appCmdHandlers.Get(i.ApplicationCommandData().Name); ok {
				err := handler(AppCmdHandlerData{C: c, S: s, I: i, Data: i.ApplicationCommandData()})
				go func() {
					if err == nil {
						return
					}
					handlerErrCh, ok := c.appCmdHandlerErrChs.Get(i.ApplicationCommandData().Name)
					if ok {
						handlerErrCh <- err
					} else if c.handlerErrCh != nil {
						c.handlerErrCh <- err
					}
				}()
			}
		case discordgo.InteractionMessageComponent:
			if handler, ok := c.msgComponentHandlers.Get(i.MessageComponentData().CustomID); ok {
				err := handler(MsgComponentHandlerData{C: c, S: s, I: i, Data: i.MessageComponentData()})
				go func() {
					if err == nil {
						return
					}
					handlerErrCh, ok := c.msgComponentHandlerErrChs.Get(i.MessageComponentData().CustomID)
					if ok {
						handlerErrCh <- err
					} else if c.handlerErrCh != nil {
						c.handlerErrCh <- err
					}
				}()
			}
		case discordgo.InteractionApplicationCommandAutocomplete:
			if handler, ok := c.appCmdAutoHandlers.Get(i.ApplicationCommandData().Name); ok {
				err := handler(AppCmdHandlerData{C: c, S: s, I: i, Data: i.ApplicationCommandData()})
				go func() {
					if err == nil {
						return
					}
					handlerErrCh, ok := c.appCmdAutoHandlerErrChs.Get(i.ApplicationCommandData().Name)
					if ok {
						handlerErrCh <- err

					} else if c.handlerErrCh != nil {
						c.handlerErrCh <- err
					}
				}()
			}
		case discordgo.InteractionModalSubmit:
			if handler, ok := c.modalSubmitHandlers.Get(i.ModalSubmitData().CustomID); ok {
				err := handler(ModalSubmitHandlerData{C: c, S: s, I: i, Data: i.ModalSubmitData()})
				go func() {
					if err == nil {
						return
					}
					handlerErrCh, ok := c.modalSubmitHandlerErrChs.Get(i.ModalSubmitData().CustomID)
					if ok {
						handlerErrCh <- err
					} else if c.handlerErrCh != nil {
						c.handlerErrCh <- err
					}
				}()
			}
		}
	})
}

func (c *Client) AddPingHandler(name string, handler PingHandler, handlerErrCh ...chan error) {
	if len(handlerErrCh) == 1 {
		c.pingHandlerErrChs.Set(name, handlerErrCh[0])
	}
	c.pingHandlers.Set(name, handler)
}

func (c *Client) AddPingHandlers(handlers map[string]PingHandler, handlerErrCh ...chan error) {
	for name, handler := range handlers {
		if len(handlerErrCh) == 1 {
			c.pingHandlerErrChs.Set(name, handlerErrCh[0])
		}
		c.pingHandlers.Set(name, handler)
	}
}

func (c *Client) AddAppCmdHandler(name string, handler AppCmdHandler, handlerErrCh ...chan error) {
	if len(handlerErrCh) == 1 {
		c.appCmdHandlerErrChs.Set(name, handlerErrCh[0])
	}
	c.appCmdHandlers.Set(name, handler)
}

func (c *Client) AddAppCmdHandlers(handlers map[string]AppCmdHandler, handlerErrCh ...chan error) {
	for name, handler := range handlers {
		if len(handlerErrCh) == 1 {
			c.appCmdHandlerErrChs.Set(name, handlerErrCh[0])
		}
		c.appCmdHandlers.Set(name, handler)
	}
}

func (c *Client) AddMsgComponentHandler(name string, handler MsgComponentHandler, handlerErrCh ...chan error) {
	if len(handlerErrCh) == 1 {
		c.msgComponentHandlerErrChs.Set(name, handlerErrCh[0])
	}
	c.msgComponentHandlers.Set(name, handler)
}

func (c *Client) AddMsgComponentHandlers(handlers map[string]MsgComponentHandler, handlerErrCh ...chan error) {
	for name, handler := range handlers {
		if len(handlerErrCh) == 1 {
			c.msgComponentHandlerErrChs.Set(name, handlerErrCh[0])
		}
		c.msgComponentHandlers.Set(name, handler)
	}
}

func (c *Client) AddAppCmdAutoHandler(name string, handler AppCmdAutoHandler, handlerErrCh ...chan error) {
	if len(handlerErrCh) == 1 {
		c.appCmdAutoHandlerErrChs.Set(name, handlerErrCh[0])
	}
	c.appCmdAutoHandlers.Set(name, handler)
}

func (c *Client) AddAppCmdAutoHandlers(handlers map[string]AppCmdAutoHandler, handlerErrCh ...chan error) {
	for name, handler := range handlers {
		if len(handlerErrCh) == 1 {
			c.appCmdAutoHandlerErrChs.Set(name, handlerErrCh[0])
		}
		c.appCmdAutoHandlers.Set(name, handler)
	}
}

func (c *Client) AddModalSubmitHandler(name string, handler ModalSubmitHandler, handlerErrCh ...chan error) {
	if len(handlerErrCh) == 1 {
		c.modalSubmitHandlerErrChs.Set(name, handlerErrCh[0])
	}
	c.modalSubmitHandlers.Set(name, handler)
}

func (c *Client) AddModalSubmitHandlers(handlers map[string]ModalSubmitHandler, handlerErrCh ...chan error) {
	for name, handler := range handlers {
		if len(handlerErrCh) == 1 {
			c.modalSubmitHandlerErrChs.Set(name, handlerErrCh[0])
		}
		c.modalSubmitHandlers.Set(name, handler)
	}
}

func (c *Client) SetHandlerErrorCh(ch chan error) {
	c.handlerErrCh = ch
}

func (c *Client) HandlerErrCh() (ch chan error) {
	return c.handlerErrCh
}

func (c *Client) StartCmds(cmds ...*discordgo.ApplicationCommand) (err error) {
	_, err = c.sess.ApplicationCommandBulkOverwrite(c.appID, "", cmds)
	return
}

func (c *Client) StartGuildCmds(guildID string, cmds ...*discordgo.ApplicationCommand) (err error) {
	_, err = c.sess.ApplicationCommandBulkOverwrite(c.appID, guildID, cmds)
	return
}

func (c *Client) Open() (err error) {
	return c.sess.Open()
}

func (c *Client) Close() {
	c.sess.Close()
}

func (c Client) Sess() (sess *discordgo.Session) {
	return c.sess
}

func (c Client) AppID() (appID string) {
	return c.appID
}

func (c Client) PingHandlers() (pingHandlers map[string]PingHandler) {
	return c.pingHandlers.Data()
}

func (c Client) AppCmdHandlers() (appCmdHandlers map[string]AppCmdHandler) {
	return c.appCmdHandlers.Data()
}

func (c Client) MsgComponentHandlers() (msgComponentHandlers map[string]MsgComponentHandler) {
	return c.msgComponentHandlers.Data()
}

func (c Client) AppCmdAutoHandlers() (appCmdAutoHandlers map[string]AppCmdAutoHandler) {
	return c.appCmdAutoHandlers.Data()
}

func (c Client) ModalSubmitHandlers() (modalSubmitHandlers map[string]ModalSubmitHandler) {
	return c.modalSubmitHandlers.Data()
}

func (c Client) PingHandlerErrChs() (pingHandlerErrChs map[string]chan error) {
	return c.pingHandlerErrChs.Data()
}

func (c Client) AppCmdHandlerErrChs() (appCmdHandlerErrChs map[string]chan error) {
	return c.appCmdHandlerErrChs.Data()
}

func (c Client) MsgComponentHandlerErrChs() (msgComponentHandlerErrChs map[string]chan error) {
	return c.msgComponentHandlerErrChs.Data()
}

func (c Client) AppCmdAutoHandlerErrChs() (appCmdAutoHandlerErrChs map[string]chan error) {
	return c.appCmdAutoHandlerErrChs.Data()
}

func (c Client) ModalSubmitHandlerErrChs() (modalSubmitHandlerErrChs map[string]chan error) {
	return c.modalSubmitHandlerErrChs.Data()
}
