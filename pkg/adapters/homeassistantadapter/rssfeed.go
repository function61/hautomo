package homeassistantadapter

import (
	"context"
	"fmt"
	"strings"

	"github.com/function61/gokit/log/logex"
	"github.com/function61/gokit/net/http/ezhttp"
	"github.com/function61/gokit/strings/stringutils"
	"github.com/function61/hautomo/pkg/homeassistant"
	"github.com/mmcdole/gofeed"
)

func makeRssFeedSensor(
	entityId string,
	feedUrl string,
	ha *homeassistant.MqttClient,
	logl *logex.Leveled,
) (*homeassistant.Entity, func(context.Context) error) {
	// need attribute topic, see comment later
	sensor := homeassistant.NewSensor(
		entityId,
		feedUrl,
		homeassistant.DeviceClassDefault,
		true)

	rssChangeDetector := &valueChangeDetector{}

	return sensor, func(ctx context.Context) error {
		feed, err := getFeedItems(ctx, feedUrl)
		if err != nil {
			return err
		}

		feedAsMarkdown := feedToMarkdown(feed, 8, 100)

		if !rssChangeDetector.Changed(feedAsMarkdown) {
			return nil
		}

		logl.Info.Printf("%s changed", entityId)

		// need to store content as an attribute, because state is capped at 256 chars
		return ha.PublishAttributes(sensor, map[string]string{
			"md": feedAsMarkdown,
		})
	}
}

// renders a Markdown table from feed items
func feedToMarkdown(feed *gofeed.Feed, maxItems int, truncate int) string {
	lines := []string{}
	line := func(l string) {
		lines = append(lines, l)
	}

	for _, item := range feed.Items {
		line(fmt.Sprintf("- [%s](%s)", stringutils.Truncate(item.Title, truncate), item.Link))

		if len(lines) >= maxItems {
			break
		}
	}

	return strings.Join(lines, "\n")
}

func getFeedItems(ctx context.Context, feedUrl string) (*gofeed.Feed, error) {
	res, err := ezhttp.Get(ctx, feedUrl)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	return gofeed.NewParser().Parse(res.Body)
}
