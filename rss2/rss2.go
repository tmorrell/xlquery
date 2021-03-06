package rss2

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"strconv"
	"strings"
)

type RSS2 struct {
	XMLName xml.Name `xml:"rss" json:"-"`
	Version string   `xml:"version,attr" json:"version"`
	// Required
	Title       string `xml:"channel>title" json:"title"`
	Link        string `xml:"channel>link" json:"link"`
	Description string `xml:"channel>description" json:"description"`
	// Optional
	PubDate  string `xml:"channel>pubDate" json:"pubDate,omitempty"`
	ItemList []Item `xml:"channel>item" json:"item,omitempty"`
}

type Item struct {
	// Optional according to Dave Winer
	Title string `xml:"title" json:"title,omitempty"`
	// Required
	Link string `xml:"link" json:"link"`
	// Optional
	Description template.HTML `xml:"description" json:"description,omitempty"`
	Content     template.HTML `xml:"encoded" json:"encoded,omitempty"`
	PubDate     string        `xml:"pubDate" json:"pubDate,omitempty"`
	Comments    string        `xml:"comments" json:"comments,omitempty"`
}

// Parse return an RSS2 document as a RSS2 structure.
func Parse(buf []byte) (*RSS2, error) {
	data := new(RSS2)
	err := xml.Unmarshal(buf, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *RSS2) channel(dataPath string) (map[string]interface{}, error) {
	results := make(map[string]interface{})
	switch {
	case strings.Compare(dataPath, ".channel") == 0:
		// package and return all the channel fields
		results[".title"] = r.Title
		results[".link"] = r.Link
		results[".description"] = r.Description
		if r.PubDate != "" {
			results[".pubDate"] = r.PubDate
		}
	case strings.HasSuffix(dataPath, ".title"):
		results[".title"] = r.Title
	case strings.HasSuffix(dataPath, ".link"):
		results[".link"] = r.Link
	case strings.HasSuffix(dataPath, ".description"):
		results[".description"] = r.Description
	case strings.HasSuffix(dataPath, ".pubDate"):
		results[".pubDate"] = r.PubDate
	default:
		return nil, fmt.Errorf("Unknown data path %s", dataPath)
	}
	return results, nil
}

type rangeExpression struct {
	first int
	last  int
}

func getRange(listLength int, exp string) *rangeExpression {
	rexp := new(rangeExpression)
	rexp.first = 0
	rexp.last = listLength - 1

	if strings.Contains(exp, "-") == true {
		nums := strings.SplitN(exp, "-", 2)
		i, err := strconv.Atoi(nums[0])
		if err == nil {
			rexp.first = i
		}
		i, err = strconv.Atoi(nums[1])
		if err == nil {
			rexp.last = i
		}
	} else {
		i, err := strconv.Atoi(exp)
		if err == nil {
			rexp.first = i
			rexp.last = i
		}
	}
	return rexp
}

func (rexp *rangeExpression) inRange(val int) bool {
	if val >= rexp.first && val <= rexp.last {
		return true
	}
	return false
}

func (r *RSS2) items(dataPath string) (map[string]interface{}, error) {
	rexp := new(rangeExpression)
	rexp.first = 0
	rexp.last = len(r.ItemList) - 1

	// Get the range expression so we know when to add it to results.
	s := strings.Index(dataPath, "[")
	e := strings.Index(dataPath, "]")
	if s >= 0 && e >= 0 {
		rexp = getRange(len(r.ItemList), dataPath[s:e])
	}

	results := make(map[string]interface{})
	switch {
	case strings.HasSuffix(dataPath, ".title") == true:
		vals := []string{}
		for i, item := range r.ItemList {
			if rexp.inRange(i) == true {
				vals = append(vals, item.Title)
			}
		}
		results["title"] = vals
	case strings.HasSuffix(dataPath, ".link") == true:
		vals := []string{}
		for i, item := range r.ItemList {
			if rexp.inRange(i) == true {
				vals = append(vals, item.Link)
			}
		}
		results["link"] = vals
		/*
			case strings.HasSuffix(dataPath, ".description") == true:
				vals := []string{}
				for i, item := range r.ItemList {
					if rexp.inRange(i) == true {
						vals = append(vals, item.Description)
					}
				}
				results["description"] = vals
			case strings.HasSuffix(dataPath, ".content") == true:
				vals := []string{}
				for i, item := range r.ItemList {
					if rexp.inRange(i) == true {
						vals = append(vals, item.Content)
					}
				}
				results["content"] = vals
		*/
	case strings.HasSuffix(dataPath, ".pubDate") == true:
		vals := []string{}
		for i, item := range r.ItemList {
			if rexp.inRange(i) == true {
				vals = append(vals, item.PubDate)
			}
		}
		results["pubDate"] = vals
	case strings.HasSuffix(dataPath, ".comments") == true:
		vals := []string{}
		for i, item := range r.ItemList {
			if rexp.inRange(i) == true {
				vals = append(vals, item.Comments)
			}
		}
		results["comments"] = vals
	}
	return results, nil
}

// Filter given an RSS2 document return all the entries matching so we can apply some sort of data path
// e.g. .version, .channel.title, .channel.link, .item[].link, .item[].guid, .item[].title, .item[].description
func (r *RSS2) Filter(dataPath string) (map[string]interface{}, error) {
	var (
		err  error
		data map[string]interface{}
	)
	result := make(map[string]interface{})
	switch {
	case strings.Compare(dataPath, ".version") == 0:
		result["version"] = r.Version
	case strings.HasPrefix(dataPath, ".channel"):
		data, err = r.channel(dataPath)
		// Merge data into results keyed' by path
		for _, val := range data {
			result[dataPath] = val
		}
	case strings.HasPrefix(dataPath, ".item[]"):
		data, err = r.items(dataPath)
		// Merge data into results keyed' by path
		for _, val := range data {
			result[dataPath] = val
		}
	default:
		return nil, fmt.Errorf("path %q not found", dataPath)
	}
	return result, err
}
