package disc

import (
	"github.com/bwmarrin/discordgo"
)

type GroupHandler struct {
	pingHandlers              map[string]PingHandler
	appCmdHandlers            map[string]AppCmdHandler
	msgComponentHandlers      map[string]MsgComponentHandler
	appCmdAutoHandlers        map[string]AppCmdAutoHandler
	modalSubmitHandlers       map[string]ModalSubmitHandler
	pingHandlerErrChs         map[string]chan error
	appCmdHandlerErrChs       map[string]chan error
	msgComponentHandlerErrChs map[string]chan error
	appCmdAutoHandlerErrChs   map[string]chan error
	modalSubmitHandlerErrChs  map[string]chan error
	errCh                     chan error
	prefixHandler             PrefixHandler
	suffixHandler             BaseHandler
	c                         *Client
}

type BaseHandler func(data BaseHandlerData) (err error)

type BaseHandlerData struct {
	C *Client
	S *discordgo.Session
	I *discordgo.InteractionCreate
}

type PrefixHandler func(data BaseHandlerData) (stop bool, err error)

type PingHandler func(data BaseHandlerData) (err error)

type AppCmdHandler func(data AppCmdHandlerData) (err error)

type AppCmdHandlerData struct {
	C    *Client
	S    *discordgo.Session
	I    *discordgo.InteractionCreate
	Data discordgo.ApplicationCommandInteractionData
}

type MsgComponentHandler func(data MsgComponentHandlerData) (err error)

type MsgComponentHandlerData struct {
	C    *Client
	S    *discordgo.Session
	I    *discordgo.InteractionCreate
	Data discordgo.MessageComponentInteractionData
}

type AppCmdAutoHandler func(data AppCmdHandlerData) (err error)

type ModalSubmitHandler func(data ModalSubmitHandlerData) (err error)

type ModalSubmitHandlerData struct {
	C    *Client
	S    *discordgo.Session
	I    *discordgo.InteractionCreate
	Data discordgo.ModalSubmitInteractionData
}

func (c *Client) NewGroupHandler() *GroupHandler {
	return &GroupHandler{
		pingHandlers:              map[string]PingHandler{},
		appCmdHandlers:            map[string]AppCmdHandler{},
		msgComponentHandlers:      map[string]MsgComponentHandler{},
		appCmdAutoHandlers:        map[string]AppCmdAutoHandler{},
		modalSubmitHandlers:       map[string]ModalSubmitHandler{},
		pingHandlerErrChs:         map[string]chan error{},
		appCmdHandlerErrChs:       map[string]chan error{},
		msgComponentHandlerErrChs: map[string]chan error{},
		appCmdAutoHandlerErrChs:   map[string]chan error{},
		modalSubmitHandlerErrChs:  map[string]chan error{},
		c:                         c,
	}
}

func (h *GroupHandler) Handle() {
	h.c.Sess().AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h.prefixHandler != nil {
			stop, err := h.prefixHandler(BaseHandlerData{C: h.c, S: s, I: i})
			if err != nil {
				if h.errCh != nil {
					h.errCh <- err
				}
				return
			}
			if stop {
				return
			}
		}
		switch i.Type {
		case discordgo.InteractionPing:
			if handler, ok := h.pingHandlers[i.ApplicationCommandData().Name]; ok {
				err := handler(BaseHandlerData{C: h.c, S: s, I: i})
				go func() {
					if err == nil {
						return
					}
					errCh, ok := h.pingHandlerErrChs[i.ApplicationCommandData().Name]
					if ok {
						errCh <- err
					} else if h.errCh != nil {
						h.errCh <- err
					}
				}()
			}
		case discordgo.InteractionApplicationCommand:
			if handler, ok := h.appCmdHandlers[i.ApplicationCommandData().Name]; ok {
				err := handler(AppCmdHandlerData{C: h.c, S: s, I: i, Data: i.ApplicationCommandData()})
				go func() {
					if err == nil {
						return
					}
					errCh, ok := h.appCmdHandlerErrChs[i.ApplicationCommandData().Name]
					if ok {
						errCh <- err
					} else if h.errCh != nil {
						h.errCh <- err
					}
				}()
			}
		case discordgo.InteractionMessageComponent:
			if handler, ok := h.msgComponentHandlers[i.MessageComponentData().CustomID]; ok {
				err := handler(MsgComponentHandlerData{C: h.c, S: s, I: i, Data: i.MessageComponentData()})
				go func() {
					if err == nil {
						return
					}

					errCh, ok := h.msgComponentHandlerErrChs[i.MessageComponentData().CustomID]
					if ok {
						errCh <- err
					} else if h.errCh != nil {
						h.errCh <- err
					}
				}()
			}
		case discordgo.InteractionApplicationCommandAutocomplete:
			if handler, ok := h.appCmdAutoHandlers[i.ApplicationCommandData().Name]; ok {
				err := handler(AppCmdHandlerData{C: h.c, S: s, I: i, Data: i.ApplicationCommandData()})
				go func() {
					if err == nil {
						return
					}
					errCh, ok := h.appCmdAutoHandlerErrChs[i.ApplicationCommandData().Name]
					if ok {
						errCh <- err
					} else if h.errCh != nil {
						h.errCh <- err
					}
				}()
			}
		case discordgo.InteractionModalSubmit:
			if handler, ok := h.modalSubmitHandlers[i.ModalSubmitData().CustomID]; ok {
				err := handler(ModalSubmitHandlerData{C: h.c, S: s, I: i, Data: i.ModalSubmitData()})
				go func() {
					if err == nil {
						return
					}
					errCh, ok := h.modalSubmitHandlerErrChs[i.ModalSubmitData().CustomID]
					if ok {
						errCh <- err
					} else if h.errCh != nil {
						h.errCh <- err
					}
				}()
			}
		}
		if h.suffixHandler != nil {
			err := h.suffixHandler(BaseHandlerData{C: h.c, S: s, I: i})
			if err != nil {
				if h.errCh != nil {
					h.errCh <- err
				}
				return
			}
		}
	})
}

func (h *GroupHandler) AddPingHandler(name string, handler PingHandler, errCh ...chan error) *GroupHandler {
	if len(errCh) == 1 {
		h.pingHandlerErrChs[name] = errCh[0]
	}
	h.pingHandlers[name] = handler
	return h
}

func (h *GroupHandler) AddPingHandlers(handlers map[string]PingHandler, errCh ...chan error) *GroupHandler {
	for name, handler := range handlers {
		if len(errCh) == 1 {
			h.pingHandlerErrChs[name] = errCh[0]
		}
		h.pingHandlers[name] = handler
	}
	return h
}

func (h *GroupHandler) AddAppCmdHandler(name string, handler AppCmdHandler, errCh ...chan error) *GroupHandler {
	if len(errCh) == 1 {
		h.appCmdHandlerErrChs[name] = errCh[0]
	}
	h.appCmdHandlers[name] = handler
	return h
}

func (h *GroupHandler) AddAppCmdHandlers(handlers map[string]AppCmdHandler, errCh ...chan error) *GroupHandler {
	for name, handler := range handlers {
		if len(errCh) == 1 {
			h.appCmdHandlerErrChs[name] = errCh[0]
		}
		h.appCmdHandlers[name] = handler
	}
	return h
}

func (h *GroupHandler) AddMsgComponentHandler(name string, handler MsgComponentHandler, errCh ...chan error) *GroupHandler {
	if len(errCh) == 1 {
		h.msgComponentHandlerErrChs[name] = errCh[0]
	}
	h.msgComponentHandlers[name] = handler
	return h
}

func (h *GroupHandler) AddMsgComponentHandlers(handlers map[string]MsgComponentHandler, errCh ...chan error) *GroupHandler {
	for name, handler := range handlers {
		if len(errCh) == 1 {
			h.msgComponentHandlerErrChs[name] = errCh[0]
		}
		h.msgComponentHandlers[name] = handler
	}
	return h
}

func (h *GroupHandler) AddAppCmdAutoHandler(name string, handler AppCmdAutoHandler, errCh ...chan error) *GroupHandler {
	if len(errCh) == 1 {
		h.appCmdAutoHandlerErrChs[name] = errCh[0]
	}
	h.appCmdAutoHandlers[name] = handler
	return h
}

func (h *GroupHandler) AddAppCmdAutoHandlers(handlers map[string]AppCmdAutoHandler, errCh ...chan error) *GroupHandler {
	for name, handler := range handlers {
		if len(errCh) == 1 {
			h.appCmdAutoHandlerErrChs[name] = errCh[0]
		}
		h.appCmdAutoHandlers[name] = handler
	}
	return h
}

func (h *GroupHandler) AddModalSubmitHandler(name string, handler ModalSubmitHandler, errCh ...chan error) *GroupHandler {
	if len(errCh) == 1 {
		h.modalSubmitHandlerErrChs[name] = errCh[0]
	}
	h.modalSubmitHandlers[name] = handler
	return h
}

func (h *GroupHandler) AddModalSubmitHandlers(handlers map[string]ModalSubmitHandler, errCh ...chan error) *GroupHandler {
	for name, handler := range handlers {
		if len(errCh) == 1 {
			h.modalSubmitHandlerErrChs[name] = errCh[0]
		}
		h.modalSubmitHandlers[name] = handler
	}
	return h
}

func (h *GroupHandler) SetErrorCh(ch chan error) *GroupHandler {
	h.errCh = ch
	return h
}

func (h *GroupHandler) SetPrefixHandler(handler PrefixHandler) *GroupHandler {
	h.prefixHandler = handler
	return h
}

func (h *GroupHandler) SetSuffixHandler(handler BaseHandler) *GroupHandler {
	h.suffixHandler = handler
	return h
}

func (h GroupHandler) PingHandlers() map[string]PingHandler {
	return h.pingHandlers
}

func (h GroupHandler) AppCmdHandlers() map[string]AppCmdHandler {
	return h.appCmdHandlers
}

func (h GroupHandler) MsgComponentHandlers() map[string]MsgComponentHandler {
	return h.msgComponentHandlers
}

func (h GroupHandler) AppCmdAutoHandlers() map[string]AppCmdAutoHandler {
	return h.appCmdAutoHandlers
}

func (h GroupHandler) ModalSubmitHandlers() map[string]ModalSubmitHandler {
	return h.modalSubmitHandlers
}

func (h GroupHandler) PingHandlerErrorChs() map[string]chan error {
	return h.pingHandlerErrChs
}

func (h GroupHandler) AppCmdHandlerErrorChs() map[string]chan error {
	return h.appCmdHandlerErrChs
}

func (h GroupHandler) MsgComponentHandlerErrorChs() map[string]chan error {
	return h.msgComponentHandlerErrChs
}

func (h GroupHandler) AppCmdAutoHandlerErrorChs() map[string]chan error {
	return h.appCmdAutoHandlerErrChs
}

func (h GroupHandler) ModalSubmitHandlerErrorChs() map[string]chan error {
	return h.modalSubmitHandlerErrChs
}

func (h GroupHandler) ErrorCh() chan error {
	return h.errCh
}

func (h GroupHandler) PrefixHandler() PrefixHandler {
	return h.prefixHandler
}

func (h GroupHandler) SuffixHandler() BaseHandler {
	return h.suffixHandler
}

func (h GroupHandler) C() *Client {
	return h.c
}
