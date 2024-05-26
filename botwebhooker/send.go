package botwebhooker

func (c *Client) sendWebhookHandler() {
	for msg := range c.sendWebhookCh {
		_, err := c.dc.Sess().ChannelMessageSendComplex(msg.cID, ConvertWebhookToMessage(msg.webhook))
		if msg.wg != nil {
			msg.wg.Done()
		}
		if err != nil {
			if c.errCh != nil {
				c.errCh <- err
			}
		}
	}
}
