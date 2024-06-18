package disc

import (
	"math"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
)

type Paginator struct {
	page     int
	perPage  int
	items    []any
	customID string
	*PaginatorBuilder
}

type PaginatorBuilder struct {
	Client            *Client
	URL               string
	URLFunc           func(*Paginator) string
	Type              discordgo.EmbedType
	TypeFunc          func(*Paginator) discordgo.EmbedType
	TitleFunc         func(*Paginator) string
	Title             string
	DescFunc          func(*Paginator) string
	Desc              string
	Timestamp         time.Time
	TimestampFunc     func(*Paginator) time.Time
	Footer            *discordgo.MessageEmbedFooter
	FooterFunc        func(*Paginator) *discordgo.MessageEmbedFooter
	Image             *discordgo.MessageEmbedImage
	ImageFunc         func(*Paginator) *discordgo.MessageEmbedImage
	Thumbnail         *discordgo.MessageEmbedThumbnail
	ThumbnailFunc     func(*Paginator) *discordgo.MessageEmbedThumbnail
	Video             *discordgo.MessageEmbedVideo
	VideoFunc         func(*Paginator) *discordgo.MessageEmbedVideo
	Provider          *discordgo.MessageEmbedProvider
	ProviderFunc      func(*Paginator) *discordgo.MessageEmbedProvider
	Author            *discordgo.MessageEmbedAuthor
	AuthorFunc        func(*Paginator) *discordgo.MessageEmbedAuthor
	Fields            []*discordgo.MessageEmbedField
	FieldsFunc        func(*Paginator) []*discordgo.MessageEmbedField
	Components        []discordgo.MessageComponent
	ComponentsFunc    func(*Paginator) []discordgo.MessageComponent
	EphemeralResponse bool
	InitialItems      []any
	OnPage            func(*Paginator) error
	OnPageErrResp     *discordgo.InteractionResponseData
	OnPageErrRespFunc func(*Paginator, error) *discordgo.InteractionResponseData
	OnResp            func(*Paginator) error
	OnRespErrResp     *discordgo.InteractionResponseData
	OnRespErrRespFunc func(*Paginator, error) *discordgo.InteractionResponseData
}

func (c *Client) NewPaginatorBuilder() *PaginatorBuilder {
	return &PaginatorBuilder{
		Client: c,
	}
}

func (b *PaginatorBuilder) Build(perPage int) (p *Paginator, msgComponentHandlers map[string]MsgComponentHandler) {
	p = &Paginator{
		page:             1,
		perPage:          perPage,
		items:            b.InitialItems,
		customID:         uuid.New().String(),
		PaginatorBuilder: b,
	}
	onPage := func(s *discordgo.Session, i *discordgo.InteractionCreate) (sent bool, err error) {
		if p.OnPage != nil {
			err := p.OnPage(p)
			if err != nil {
				if p.OnPageErrRespFunc != nil {
					return true, s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: p.OnPageErrRespFunc(p, err),
					})
				}
				if p.OnPageErrResp != nil {
					return true, s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: p.OnPageErrResp,
					})
				}
				return true, s.InteractionRespond(i.Interaction, p.UpdateResponse())
			}
		}
		return
	}
	return p, map[string]MsgComponentHandler{
		p.customID + "-prev": func(data MsgComponentHandlerData) error {
			sent, err := onPage(data.S, data.I)
			if err != nil {
				return err
			}
			if sent {
				return nil
			}
			p.page--
			return data.S.InteractionRespond(data.I.Interaction, p.UpdateResponse())
		},
		p.customID + "-next": func(data MsgComponentHandlerData) error {
			sent, err := onPage(data.S, data.I)
			if err != nil {
				return err
			}
			if sent {
				return nil
			}
			p.page++
			return data.S.InteractionRespond(data.I.Interaction, p.UpdateResponse())
		},
	}
}

func (p Paginator) Items() []any {
	return p.items
}

func (p *Paginator) SetItems(items []any) {
	p.items = items
}

func (p Paginator) Page() int {
	return p.page
}

func (p Paginator) PerPage() int {
	return p.perPage
}

func (p Paginator) LastPage() int {
	return int(math.Ceil(float64(float64(len(p.items)) / float64(p.perPage))))
}

func (p Paginator) CurPageItems() []any {
	start := (p.page - 1) * p.perPage
	if start > len(p.items) {
		start = len(p.items)
	}
	end := start + p.perPage
	if end > len(p.items) {
		end = len(p.items)
	}

	return p.items[start:end]
}

func (p Paginator) CurPageIdxs() (idxs []int) {
	start := (p.page - 1) * p.perPage
	end := start + p.perPage
	if end > len(p.items) {
		end = len(p.items)
	}
	for i := start; i < end; i++ {
		idxs = append(idxs, i)
	}
	return
}

func (p *Paginator) Response() *discordgo.InteractionResponse {
	if p.OnResp != nil {
		err := p.OnResp(p)
		if err != nil {
			if p.OnRespErrRespFunc != nil {
				return &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: p.OnRespErrRespFunc(p, err),
				}
			}
			if p.OnRespErrResp != nil {
				return &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: p.OnRespErrResp,
				}
			}
			return p.UpdateResponse()
		}
	}
	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					URL:         p.getURL(),
					Type:        p.getType(),
					Title:       p.getTitle(),
					Description: p.getDesc(),
					Timestamp:   p.getTimestamp(),
					Footer:      p.getFooter(),
					Image:       p.getImage(),
					Thumbnail:   p.getThumbnail(),
					Video:       p.getVideo(),
					Provider:    p.getProvider(),
					Author:      p.getAuthor(),
					Fields:      p.getFields(),
				},
			},
			Components: p.getComponents(),
			Flags:      p.getFlags(),
		},
	}
}

func (p *Paginator) UpdateResponse() *discordgo.InteractionResponse {
	if p.OnResp != nil {
		err := p.OnResp(p)
		if err != nil {
			if p.OnRespErrRespFunc != nil {
				return &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: p.OnRespErrRespFunc(p, err),
				}
			}
			if p.OnRespErrResp != nil {
				return &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: p.OnRespErrResp,
				}
			}
			return p.UpdateResponse()
		}
	}
	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					URL:         p.getURL(),
					Type:        p.getType(),
					Title:       p.getTitle(),
					Description: p.getDesc(),
					Timestamp:   p.getTimestamp(),
					Footer:      p.getFooter(),
					Image:       p.getImage(),
					Thumbnail:   p.getThumbnail(),
					Video:       p.getVideo(),
					Provider:    p.getProvider(),
					Author:      p.getAuthor(),
					Fields:      p.getFields(),
				},
			},
			Components: p.getComponents(),
			Flags:      p.getFlags(),
		},
	}
}

func (p Paginator) getURL() string {
	if p.URLFunc != nil {
		return p.URLFunc(&p)
	}
	return p.URL
}

func (p *Paginator) getType() discordgo.EmbedType {
	if p.TypeFunc != nil {
		return p.TypeFunc(p)
	}
	return p.Type
}

func (p *Paginator) getTimestamp() string {

	if p.TimestampFunc != nil {
		if p.TimestampFunc(p).IsZero() {
			return ""
		}
		return p.TimestampFunc(p).UTC().Format("2006-01-02T15:04:05.999Z")
	}
	if p.Timestamp.IsZero() {
		return ""
	}
	return p.Timestamp.UTC().Format("2006-01-02T15:04:05.999Z")
}

func (p *Paginator) getFooter() *discordgo.MessageEmbedFooter {
	if p.FooterFunc != nil {
		return p.FooterFunc(p)
	}
	return p.Footer
}

func (p *Paginator) getImage() *discordgo.MessageEmbedImage {
	if p.ImageFunc != nil {
		return p.ImageFunc(p)
	}
	return p.Image
}

func (p *Paginator) getThumbnail() *discordgo.MessageEmbedThumbnail {
	if p.ThumbnailFunc != nil {
		return p.ThumbnailFunc(p)
	}
	return p.Thumbnail
}

func (p *Paginator) getVideo() *discordgo.MessageEmbedVideo {
	if p.VideoFunc != nil {
		return p.VideoFunc(p)
	}
	return p.Video
}

func (p *Paginator) getProvider() *discordgo.MessageEmbedProvider {
	if p.ProviderFunc != nil {
		return p.ProviderFunc(p)
	}
	return p.Provider
}

func (p *Paginator) getAuthor() *discordgo.MessageEmbedAuthor {
	if p.AuthorFunc != nil {
		return p.AuthorFunc(p)
	}
	return p.Author
}

func (p *Paginator) getFields() []*discordgo.MessageEmbedField {
	if p.FieldsFunc != nil {
		return p.FieldsFunc(p)
	}
	return p.Fields
}

func (p *Paginator) getTitle() string {
	if p.TitleFunc != nil {
		return p.TitleFunc(p)
	}
	return p.Title
}

func (p *Paginator) getDesc() string {
	if p.DescFunc != nil {
		return p.DescFunc(p)
	}
	return p.Desc
}

func (p *Paginator) getComponents() []discordgo.MessageComponent {
	baseComponents := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Prev",
					Style:    discordgo.PrimaryButton,
					CustomID: p.customID + "-prev",
					Disabled: p.page == 1,
				},
				discordgo.Button{
					Label:    "Next",
					Style:    discordgo.PrimaryButton,
					CustomID: p.customID + "-next",
					Disabled: p.page == p.LastPage() || p.LastPage() == 0,
				},
			},
		},
	}
	if p.ComponentsFunc != nil {
		return append(p.ComponentsFunc(p), baseComponents...)
	}
	return append(p.Components, baseComponents...)
}

func (p *Paginator) getFlags() discordgo.MessageFlags {
	if p.EphemeralResponse {
		return 1 << 6
	}
	return 0
}

func (p *PaginatorBuilder) SetTitleFunc(titleFunc func(*Paginator) string) *PaginatorBuilder {
	p.TitleFunc = titleFunc
	return p
}

func (p *PaginatorBuilder) SetTitle(title string) *PaginatorBuilder {
	p.Title = title
	return p
}

func (p *PaginatorBuilder) SetDescFunc(descFunc func(*Paginator) string) *PaginatorBuilder {
	p.DescFunc = descFunc
	return p
}

func (p *PaginatorBuilder) SetDesc(desc string) *PaginatorBuilder {
	p.Desc = desc
	return p
}

func (p *PaginatorBuilder) UseEphemeralResponse() *PaginatorBuilder {
	p.EphemeralResponse = true
	return p
}

func (p *PaginatorBuilder) SetInitialItems(items []any) *PaginatorBuilder {
	p.InitialItems = items
	return p
}

func (p *PaginatorBuilder) SetURLFunc(urlFunc func(*Paginator) string) *PaginatorBuilder {
	p.URLFunc = urlFunc
	return p
}

func (p *PaginatorBuilder) SetURL(url string) *PaginatorBuilder {
	p.URL = url
	return p
}

func (p *PaginatorBuilder) SetTypeFunc(t func(*Paginator) discordgo.EmbedType) *PaginatorBuilder {
	p.TypeFunc = t
	return p
}

func (p *PaginatorBuilder) SetType(t discordgo.EmbedType) *PaginatorBuilder {
	p.Type = t
	return p
}

func (p *PaginatorBuilder) SetTimestampFunc(t func(*Paginator) time.Time) *PaginatorBuilder {
	p.TimestampFunc = t
	return p
}

func (p *PaginatorBuilder) SetTimestamp(t time.Time) *PaginatorBuilder {
	p.Timestamp = t
	return p
}

func (p *PaginatorBuilder) SetFooterFunc(f func(*Paginator) *discordgo.MessageEmbedFooter) *PaginatorBuilder {
	p.FooterFunc = f
	return p
}

func (p *PaginatorBuilder) SetFooter(f *discordgo.MessageEmbedFooter) *PaginatorBuilder {
	p.Footer = f
	return p
}

func (p *PaginatorBuilder) SetImageFunc(i func(*Paginator) *discordgo.MessageEmbedImage) *PaginatorBuilder {
	p.ImageFunc = i
	return p
}

func (p *PaginatorBuilder) SetImage(i *discordgo.MessageEmbedImage) *PaginatorBuilder {
	p.Image = i
	return p
}

func (p *PaginatorBuilder) SetThumbnailFunc(t func(*Paginator) *discordgo.MessageEmbedThumbnail) *PaginatorBuilder {
	p.ThumbnailFunc = t
	return p
}

func (p *PaginatorBuilder) SetThumbnail(t *discordgo.MessageEmbedThumbnail) *PaginatorBuilder {
	p.Thumbnail = t
	return p
}

func (p *PaginatorBuilder) SetVideoFunc(v func(*Paginator) *discordgo.MessageEmbedVideo) *PaginatorBuilder {
	p.VideoFunc = v
	return p
}

func (p *PaginatorBuilder) SetVideo(v *discordgo.MessageEmbedVideo) *PaginatorBuilder {
	p.Video = v
	return p
}

func (p *PaginatorBuilder) SetProviderFunc(pr func(*Paginator) *discordgo.MessageEmbedProvider) *PaginatorBuilder {
	p.ProviderFunc = pr
	return p
}

func (p *PaginatorBuilder) SetProvider(pr *discordgo.MessageEmbedProvider) *PaginatorBuilder {
	p.Provider = pr
	return p
}

func (p *PaginatorBuilder) SetAuthorFunc(a func(*Paginator) *discordgo.MessageEmbedAuthor) *PaginatorBuilder {
	p.AuthorFunc = a
	return p
}

func (p *PaginatorBuilder) SetAuthor(a *discordgo.MessageEmbedAuthor) *PaginatorBuilder {
	p.Author = a
	return p
}

func (p *PaginatorBuilder) SetFieldsFunc(f func(*Paginator) []*discordgo.MessageEmbedField) *PaginatorBuilder {
	p.FieldsFunc = f
	return p
}

func (p *PaginatorBuilder) SetFields(f []*discordgo.MessageEmbedField) *PaginatorBuilder {
	p.Fields = f
	return p
}

func (p *PaginatorBuilder) SetComponentsFunc(c func(*Paginator) []discordgo.MessageComponent) *PaginatorBuilder {
	p.ComponentsFunc = c
	return p
}

func (p *PaginatorBuilder) SetComponents(c []discordgo.MessageComponent) *PaginatorBuilder {
	p.Components = c
	return p
}

func (p *PaginatorBuilder) SetOnPage(onPage func(*Paginator) error) *PaginatorBuilder {
	p.OnPage = onPage
	return p
}

func (p *PaginatorBuilder) SetOnPageErrResp(onPageErrResp *discordgo.InteractionResponseData) *PaginatorBuilder {
	p.OnPageErrResp = onPageErrResp
	return p
}

func (p *PaginatorBuilder) SetOnPageErrRespFunc(onPageErrResp func(*Paginator, error) *discordgo.InteractionResponseData) *PaginatorBuilder {
	p.OnPageErrRespFunc = onPageErrResp
	return p
}

func (p *PaginatorBuilder) SetOnResp(onResp func(*Paginator) error) *PaginatorBuilder {
	p.OnResp = onResp
	return p
}

func (p *PaginatorBuilder) SetOnRespErrResp(onRespErrResp *discordgo.InteractionResponseData) *PaginatorBuilder {
	p.OnRespErrResp = onRespErrResp
	return p
}

func (p *PaginatorBuilder) SetOnRespErrRespFunc(onRespErrResp func(*Paginator, error) *discordgo.InteractionResponseData) *PaginatorBuilder {
	p.OnRespErrRespFunc = onRespErrResp
	return p
}
