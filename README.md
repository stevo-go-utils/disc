# disc
A simplified wrapper for [discordgo](https://github.com/bwmarrin/discordgo). Focuses on easily handling interactions to help prevent any simple mistakes in your code.

## Creating a client
```go
discClient, err := disc.NewClient(os.Getenv("BOT_TOKEN"), os.Getenv("BOT_APP_ID"))
if err != nil {
    panic(err)
}
// Must open the connection similar to discordgo, otherwise bot will not function.
err = c.Open()
if err != nil {
    panic(err)
}
defer c.Close() // Close the connection upon exiting
```

## Spawning Commands
### Customize The Command
You can import the same commands you previously created with discordgo.
#### Example Cmd
```go
fooCmd := &discordgo.ApplicationCommand{
    Name:        "foo",
    Description: "bar",
    ...
}
```
### Spawn The Command
```go
err := discClient.StartCmds(fooCmd)
if err != nil {
    panic(err)
}
```
### Spawn The Command For A Guild
```go
err := discClient.StartGuildCmds(os.Getenv("GUILD_ID"), fooCmd)
if err != nil {
    panic(err)
}
```
## Handling Commands
Using disc's handler functions adding handlers for commands is simple. There are two methods to add an AppCmdHandler: directly adding to the client or creating a group handler. Here's both methods.
### Using The Client Handler
#### Start The Client Handler
This will add a handler preset to handle any handlers you add to the client. If you call this function multiple times it will spawn duplicate handlers, so only call it once.
```go
discClient.Handle()
```
#### Add The Hander
After calling this function the handler will be added to the client's main handler. You can call this function before or after calling `discClient.Handle()`. Just know the command's interactions will not be processed until the AppCmdHandler is added.
```go
discClient.AddAppCmdHandler(
    /* Name Of The Command To Handle */ "foo", 
    /* Handler Function */ func(data disc.AppCmdHandlerData) (err error) {
        s := data.S
        i := data.I
        return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: "bar",
            },
        })
    })
```
### Using The Group Handler
We provide the same data that was used for the previous method, but the group handler can provide specific handling for the subset of handlers you provide.
```go
discClient.NewGroupHandler().AddAppCmdHandler("foo", func(data disc.AppCmdHandlerData) (err error) {
    s := data.S
    i := data.I
    return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: "bar",
        },
    })
}).Handle()
```
Until you call the `.Handle()` method on the GroupHandler builder the provided handlers will have no functionality.

## Anchors
Anchors are a functionality built for channels that serve a single purpose of displaying a message by the bot. Such as, TOS and rule or a verify button. Specify the channel where the message should be anchored and customize how you want the message to be displayed.
### Creating An Anchor
```go
err := discClient.Anchor(os.Getenv("CHANNEL_ID"), &discordgo.MessageSend{
    Embed: &discordgo.MessageEmbed{
        Title:       "Test",
        Description: "Test",
    },
    Components: []discordgo.MessageComponent{
        discordgo.ActionsRow{
            Components: []discordgo.MessageComponent{
                discordgo.Button{
                    Label:    "Test",
                    Style:    discordgo.PrimaryButton,
                    CustomID: "test",
                },
            },
        },
    },
}, disc.ForceClearAnchorOpt())
if err != nil {
    panic(err)
}
```
### Options
#### Force Clear
The `ForceClearAnchorOpt()` will delete ALL messages within that channel including ones the bot previously sent. This can be useful when you want to easily update an anchor.
#### Max Available Messages
The `MaxAllowedMessagesAnchorOpt(x)` will keep x maxium messages left when deleting previous messages from a channel. This can be useful when posting multiple anchors in a channel.

## Paginator
### Creating A Paginator
