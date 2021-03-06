package wporg

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
)

const (
	wpChangelogURL = "https://%s.trac.wordpress.org/log/?verbose=on&mode=follow_copy&format=changelog&rev=%d&limit=%d"
)

var (
	regexChangelog = regexp.MustCompile(`(?s)\[(\d+?)\].+?\* (.+?)[\/| ]`)
)

// GetChangeLog fetches a list of updated Plugins/Themes from between the provided revisions
func (c *Client) GetChangeLog(dir string, current, latest int) ([][]string, error) {
	var list [][]string
	diff := latest - current
	limit := current + 100

	for limit < latest && diff > 100 {
		URL := fmt.Sprintf(wpChangelogURL, dir, limit, 100)
		items, err := c.doChangeLog(URL, current)
		if err != nil {
			return list, err
		}
		list = append(list, items...)
		limit += 100
	}

	// We are less than 100 updates behind, make one request
	URL := fmt.Sprintf(wpChangelogURL, dir, latest, 100)
	items, err := c.doChangeLog(URL, current)
	if err != nil {
		return list, err
	}
	list = append(list, items...)

	// Reverse array so lowest revision is first
	list = reverseList(list)

	return list, err
}

func (c *Client) doChangeLog(URL string, revision int) ([][]string, error) {
	var list [][]string

	// Make the Request
	resp, err := c.getRequest(URL)
	if err != nil {
		return list, err
	}

	// Drain body and check Close error
	defer drainAndClose(resp.Body, &err)
	bytes, err := ioutil.ReadAll(resp.Body)

	matches := regexChangelog.FindAllStringSubmatch(string(bytes), -1)

	// Get the desired substring match
	for _, match := range matches {
		matchRev, err := strconv.Atoi(match[1])
		if err != nil {
			continue
		}
		if matchRev > revision {
			list = append(list, []string{match[2], match[1]})
		}
	}

	return list, err
}

// reverseList reverses the array
func reverseList(list [][]string) [][]string {
	last := len(list) - 1
	for i := 0; i < len(list)/2; i++ {
		list[i], list[last-i] = list[last-i], list[i]
	}
	return list
}
